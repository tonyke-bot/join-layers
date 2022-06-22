package generate

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/akamensky/argparse"
	tm "github.com/buger/goterm"
	json "github.com/json-iterator/go"
	"image"
	"image/png"
	"join-layers/config"
	"join-layers/util"
	"math"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/image/draw"
)

type Args struct {
	ConfigFile   *string
	OutputFolder *string
	LayerFolder  *string
	Concurrency  *int
}

var cfg *config.Config
var args = Args{}

func Setup(parser *argparse.Parser) *argparse.Command {
	cmd := parser.NewCommand("generate", "generate assets")

	args.ConfigFile = cmd.String("c", "config", &argparse.Options{
		Required: false,
		Validate: util.ValidateStringArgs,
		Help:     "path to the config file. default: config.yaml",
		Default:  "config.yaml",
	})

	args.LayerFolder = cmd.String("l", "layers", &argparse.Options{
		Required: false,
		Validate: util.ValidateStringArgs,
		Help:     "path to the layers folder. default: ./layers",
		Default:  "layers",
	})

	args.OutputFolder = cmd.String("o", "output", &argparse.Options{
		Required: false,
		Validate: util.ValidateStringArgs,
		Help:     "path to the output folder. default: config.yaml",
		Default:  "output",
	})

	args.Concurrency = cmd.Int("p", "parallel", &argparse.Options{
		Required: false,
		Help:     fmt.Sprintf("concurrency of image generation. default: %v", runtime.NumCPU()),
		Default:  runtime.NumCPU(),
	})

	return cmd
}

func getImagesForLayer(imageCache map[string]Image, layerFolder, layerName string) []*Image {
	files, err := os.ReadDir(layerFolder)
	if err != nil {
		fmt.Printf("Fail to enumerate folder %v: %v\r\n", layerFolder, err)
		os.Exit(-1)
	}

	images := make([]*Image, 0)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.ToLower(path.Ext(file.Name())) != ".png" {
			continue
		}

		fullPath := path.Join(layerFolder, file.Name())
		if cache, ok := imageCache[fullPath]; ok {
			copyCache := cache
			copyCache.LayerName = layerName
			images = append(images, &copyCache)
			continue
		}

		fmt.Printf(tm.ResetLine(fmt.Sprintf("Loading file: %v", fullPath)))

		baseName := util.WithoutExt(path.Base(file.Name()))
		separatorPos := strings.LastIndexByte(baseName, '#')
		if separatorPos == -1 {
			fmt.Printf("File %v doesn't have rarity info\r\n", fullPath)
			os.Exit(-1)
		}

		name := baseName[:separatorPos]
		rarity := baseName[separatorPos+1:]

		if name == "" {
			fmt.Printf("File %v has blank name\r\n", fullPath)
			os.Exit(-1)
		}

		rarityFloat, err := strconv.ParseFloat(rarity, 32)
		if err != nil {
			fmt.Printf("File %v has invalid rarity info: %v\r\n", fullPath, err)
			os.Exit(-1)
		} else if rarityFloat == 0 {
			fmt.Printf("File %v has invalid rarity info\r\n", fullPath)
			os.Exit(-1)
		}

		imageData, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("Fail to load %v: %v\r\n", fullPath, err)
			os.Exit(-1)
		}

		imageObject, err := png.Decode(bytes.NewReader(imageData))
		if err != nil {
			fmt.Printf("Fail to decode %v: %v\r\n", fullPath, err)
			os.Exit(-1)
		}

		img := Image{
			Name:      name,
			LayerName: layerName,
			FullPath:  fullPath,
			Rarity:    rarityFloat,
			Obj:       imageObject,
			Hash:      util.SHA1Hash(imageData),
		}

		images = append(images, &img)
		imageCache[fullPath] = img
	}

	fmt.Printf(tm.ResetLine(""))

	return images
}

func calculateDNA(images []*Image) string {
	chunks := make([][]byte, len(images))

	for idx, img := range images {
		chunks[idx] = img.Hash
	}

	return hex.EncodeToString(util.SHA1Hash(chunks...))
}

type Item struct {
	NamePrefix string
	Layers     []*Image

	ID        uint
	Metadata  []byte
	ImageData []byte
}

func (i *Item) MergeLayers(width, height uint) error {
	canvas := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	for _, layer := range i.Layers {
		if uint(layer.Obj.Bounds().Max.Y) == height && uint(layer.Obj.Bounds().Max.X) == width {
			draw.Draw(canvas, layer.Obj.Bounds(), layer.Obj, image.Point{}, draw.Over)
		} else {
			draw.NearestNeighbor.Scale(canvas, canvas.Bounds(), layer.Obj, layer.Obj.Bounds(), draw.Over, &draw.Options{})
		}
	}

	buffer := bytes.Buffer{}
	err := png.Encode(&buffer, canvas)
	if err != nil {
		return err
	}

	i.ImageData = buffer.Bytes()
	return nil
}

func processGenerateWorker(wg *sync.WaitGroup, chItems <-chan Item, completedItems chan<- Item, counter *uint32) {
	defer wg.Done()

	for {
		select {
		case item := <-chItems:
			err := item.MergeLayers(cfg.Width, cfg.Width)
			if err != nil {
				fmt.Printf("Fail to generate file: %v\r\n", err)
				os.Exit(-1)
			}

			attributes := make([]map[string]any, 0, len(item.Layers))
			for _, layer := range item.Layers {
				attributes = append(attributes, map[string]any{
					"trait_type": layer.LayerName,
					"value":      layer.Name,
				})
			}

			imageURL, _ := cfg.TypedBaseURI.Parse(fmt.Sprintf("%v.png", item.ID))

			metadata := map[string]any{
				"name":       fmt.Sprintf("%s #%v", item.NamePrefix, item.ID),
				"id":         item.ID,
				"image":      imageURL.String(),
				"attributes": attributes,
			}

			if cfg.IsSolana {
				metadata["properties"] = map[string]any{
					"files": []any{map[string]any{
						"uri":  imageURL.String(),
						"type": "image/png",
					}},
					"category": "image",
					"creators": []any{map[string]any{
						"address": *cfg.CreatorAddress,
						"share":   100,
					}},
				}
			}

			for k, data := range cfg.AdditionalData {
				metadata[k] = data
			}

			item.Metadata, err = json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				fmt.Printf("Fail to marshal metadata: %v\r\n", err)
				os.Exit(-1)
			}

			atomic.AddUint32(counter, 1)
			completedItems <- item
		}
	}
}

func processFileWorker(items <-chan Item, meter *util.ProgressMeter) {
	baseJSONDir := path.Join(*args.OutputFolder, JSONFolderName)
	baseImagesDir := path.Join(*args.OutputFolder, ImagesFolderName)

	for {
		select {
		case file := <-items:
			var err error
			metadataFilePath := path.Join(baseJSONDir, fmt.Sprintf("%v.json", file.ID))
			err = os.WriteFile(metadataFilePath, file.Metadata, 0644)
			if err != nil {
				fmt.Printf("Fail to write %v: %v\r\n", metadataFilePath, err)
				os.Exit(-1)
			}

			imageFilePath := path.Join(baseImagesDir, fmt.Sprintf("%v.png", file.ID))
			err = os.WriteFile(imageFilePath, file.ImageData, 0644)
			if err != nil {
				fmt.Printf("Fail to write %v: %v\r\n", imageFilePath, err)
				os.Exit(-1)
			}

			meter.Log()
		}
	}
}

func ensureFolders() {
	foldersToCreate := []string{
		path.Join(*args.OutputFolder, JSONFolderName),
		path.Join(*args.OutputFolder, ImagesFolderName),
	}

	for _, folder := range foldersToCreate {
		if fInfo, err := os.Stat(folder); os.IsNotExist(err) {
			err = os.MkdirAll(folder, os.ModePerm)
			if err != nil {
				fmt.Printf("Fail to create folder %v: %v\r\n", folder, err)
				os.Exit(-1)
			}
		} else if err != nil {
			fmt.Printf("Fail to get info of %v: %v\r\n", folder, err)
			os.Exit(-1)
		} else if !fInfo.IsDir() {
			fmt.Printf("%v is not a folder\r\n", folder)
			os.Exit(-1)
		}
	}

	fmt.Println("Metadata folder:", foldersToCreate[0])
	fmt.Println("Image folder:", foldersToCreate[1])
}

func composeItems(cfg *config.Config) []*Item {
	rand.Seed(time.Now().UnixMilli())

	items := make([]*Item, 0)
	imageCache := make(map[string]Image)

	for layerSetIdx, layerSet := range cfg.TraitSets {
		expandedLayers := make([][]*Image, len(layerSet.TraitsOrder))

		for layerIdx, layer := range layerSet.TraitsOrder {
			layerFolder := path.Join(*args.LayerFolder, layer.Name)
			layerName := layer.Name
			if layer.DisplayName != nil {
				layerName = *layer.DisplayName
			}
			images := getImagesForLayer(imageCache, layerFolder, layerName)

			expandedLayer := make([]*Image, 0, layerSet.Size)

			raritySum := float64(0)
			for _, img := range images {
				raritySum += img.Rarity
			}

			distributed := uint(0)
			distribution := make([]uint, len(images))
			for idx, img := range images {
				distribution[idx] = uint(math.Trunc(float64(layerSet.Size) * img.Rarity / raritySum))
				distributed += distribution[idx]

				for i := uint(0); i < distribution[idx]; i++ {
					expandedLayer = append(expandedLayer, img)
				}
			}

			for ; distributed < layerSet.Size; distributed++ {
				imageToPick := uint(rand.Uint32()) % uint(len(images))
				expandedLayer = append(expandedLayer, images[imageToPick])
			}

			rand.Shuffle(len(expandedLayer), func(i, j int) {
				expandedLayer[i], expandedLayer[j] = expandedLayer[j], expandedLayer[i]
			})

			expandedLayers[layerIdx] = expandedLayer
		}

		maxRetry := 20
		for ; cfg.CheckDuplication && maxRetry > 0; maxRetry-- {
			dnaDuplication := map[string]bool{}
			noDuplication := true

			for i := uint(0); i < layerSet.Size && maxRetry > 0; i++ {
				layers := make([]*Image, len(layerSet.TraitsOrder))

				for layerIdx := 0; layerIdx < len(layerSet.TraitsOrder); layerIdx++ {
					layers[layerIdx] = expandedLayers[layerIdx][i]
				}

				if dna := calculateDNA(layers); !dnaDuplication[dna] {
					dnaDuplication[dna] = true
					continue
				}

				noDuplication = false

				// When duplication happens, all layers in this layer set will be shuffled again and retry until there
				// is no duplication or exceeds max allowed retry times.

				fmt.Printf("DNA Duplicated, Retrying: %v\n", maxRetry)

				for layerIdx := 0; layerIdx < len(layerSet.TraitsOrder); layerIdx++ {
					layer := expandedLayers[layerIdx]
					rand.Shuffle(len(layer), func(i, j int) {
						layer[i], layer[j] = layer[j], layer[i]
					})
				}

				break
			}

			if noDuplication {
				break
			}
		}

		if maxRetry <= 0 {
			fmt.Printf("Too much duplication! Please try to reduce size of Layer Set #%v\r\n", layerSetIdx)
			os.Exit(-1)
		}

		name := cfg.Name
		if layerSet.Name != nil {
			name = *layerSet.Name
		}

		for i := uint(0); i < layerSet.Size; i++ {
			layers := make([]*Image, len(layerSet.TraitsOrder))

			for layerIdx := range layerSet.TraitsOrder {
				layers[layerIdx] = expandedLayers[layerIdx][i]
			}

			items = append(items, &Item{
				NamePrefix: name,
				Layers:     layers,
			})
		}
	}

	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	return items
}

func Exec() {
	cfg = config.Load(*args.ConfigFile)

	runtime.GOMAXPROCS(*args.Concurrency)
	ensureFolders()

	items := composeItems(cfg)
	fmt.Println("Items to generate:", len(items))

	wg := &sync.WaitGroup{}
	concurrency := runtime.NumCPU()

	chPendingItems := make(chan Item, concurrency*1000)
	chCompletedItems := make(chan Item, concurrency*5)

	generatedItems := uint32(0)
	progress := util.NewProgressMeter(time.Minute*5, uint32(len(items)))

	wg.Add(concurrency)

	fmt.Println("Starting workers...")

	go processFileWorker(chCompletedItems, progress)

	for i := 0; i < concurrency; i++ {
		go processGenerateWorker(wg, chPendingItems, chCompletedItems, &generatedItems)
	}

	startTime := time.Now()

	go func() {
		fmt.Println("Adding tasks...")

		for idx, item := range items {
			item.ID = cfg.ActualStartID() + uint(idx)
			chPendingItems <- *item
		}
	}()

	refreshInterval := time.Millisecond * 100
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			msg := fmt.Sprintf(
				"Generated/Written/Total: %v/%v/%v.",
				atomic.LoadUint32(&generatedItems),
				progress.Current(),
				len(items),
			)

			if eta := progress.ETA(); eta != 0 {
				msg += " Estimated Time Left: " + eta.Round(time.Second).String()
			}

			fmt.Printf(tm.ResetLine(msg))

			if progress.Finished() {
				fmt.Println(tm.ResetLine("Finished! Time used: " + time.Now().Sub(startTime).Round(time.Millisecond).String()))
				return
			}
		}
	}
}
