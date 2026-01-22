# CloudComPy Point Cloud Processing

Batch processing scripts for automating point cloud workflows using **CloudComPy** (Python bindings for CloudCompare). Process LAS files through normal computation, Poisson Surface Reconstruction with color transfer, and save CloudCompare projects for manual mesh filtering.

**Now with a beautiful Terminal User Interface (TUI)** built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) featuring animated progress indicators and real-time feedback!

## Features

- **Interactive TUI**: Beautiful terminal interface with animations for configuring and running processing
- **Animated Progress**: Step-by-step pipeline visualization with unique spinners for each stage
- **Real-time Stats**: Watch point counts, mesh faces, and elapsed time update live
- **Batch Processing**: Automatically process multiple LAS files in a directory
- **Normal Computation**: Calculate normals using triangulation with MST orientation
- **DIP/Dip Direction**: Convert normals to scalar fields for geological analysis
- **Surface Reconstruction**: Poisson Surface Reconstruction with density output
- **Color Transfer**: Interpolate RGB colors from point cloud to mesh vertices
- **CloudCompare Projects**: Output `.bin` files ready for filtering and export
- **Clipboard Support**: Paste paths directly with `Ctrl+V`

## Requirements

- Windows 10 or 11
- Anaconda or Miniconda
- CloudComPy binaries (Python 3.11 version)
- Go 1.21+ (for building the TUI)

## Installation

### Step 1: Set Up Conda Environment

Run the setup script from Anaconda Prompt:

```batch
.\setup_cloudcompy.bat
```

This creates the `CloudComPy311` conda environment with all required packages (~10-30 minutes).

### Step 2: Download CloudComPy Binaries

1. Download from: https://www.simulation.openfields.fr/index.php/cloudcompy-downloads
2. Choose the **Python 3.11** version for Windows
3. Extract to `C:\bin\CloudComPy311`

### Step 3: Verify Installation

```batch
conda activate CloudComPy311
cd C:\bin\CloudComPy311
envCloudComPy.bat
python -c "import cloudComPy as cc; print('CloudComPy OK!')"
```

## Usage

### TUI Mode (Recommended)

Build and run the interactive terminal interface:

```batch
.\build.bat
.\cloudcompare-tui.exe
```

### TUI Screens

#### Welcome Screen
- Overview of the tool with ASCII art logo
- Press `Enter` to start

#### Configuration Screen
- **Input Directory**: Path to folder containing LAS files (supports `Ctrl+V` paste)
- **Output Directory**: Subdirectory name for output files (default: `Processed`)
- **KNN**: K-nearest neighbors for MST normal orientation (default: 6)
- **Octree Depth**: Poisson reconstruction depth (default: 11, range 8-12)
- **Samples/Node**: Samples per node parameter (default: 1.5)
- **Point Weight**: Point weight parameter (default: 2.0)
- **Boundary Type**: 0=Free, 1=Dirichlet, 2=Neumann (default: 2)
- **Summary Panel**: Shows full paths, quality setting, and LAS file count

### TUI Navigation

| Key | Action |
|-----|--------|
| `Tab` / `↓` | Next field |
| `Shift+Tab` / `↑` | Previous field |
| `Enter` | Submit / Select / Start |
| `b` | Browse for directory |
| `Ctrl+V` | Paste from clipboard |
| `Esc` | Go back |
| `q` | Quit |
| `Ctrl+C` | Cancel processing |

### Command Line Mode

Process all LAS files in the current directory:

```batch
.\run_cloudcompy.bat
```

Process files in a specific directory:

```batch
.\run_cloudcompy.bat C:\path\to\las\files
```

### Command Line Options

```
.\run_cloudcompy.bat [input_dir] [options]

Options:
  --output-dir NAME       Output subdirectory name (default: Processed)
  --knn N                 K-nearest neighbors for MST orientation (default: 6)
  --octree-depth N        Octree depth for Poisson reconstruction (default: 11)
  --samples-per-node F    Samples per node (default: 1.5)
  --point-weight F        Point weight for interpolation (default: 2.0)
  --boundary-type N       0=Free, 1=Dirichlet, 2=Neumann (default: 2)
  --quiet                 Suppress progress output
```

### Examples

```batch
# Fast processing with lower detail
.\run_cloudcompy.bat --octree-depth 8

# High quality reconstruction
.\run_cloudcompy.bat --octree-depth 12 --samples-per-node 1.0

# Process specific folder
.\run_cloudcompy.bat D:\PointClouds --output-dir Results
```

### Octree Depth Guide

| Depth | Speed    | Detail | Memory  | Use Case                    |
|-------|----------|--------|---------|-----------------------------|
| 8     | Fast     | Low    | ~1 GB   | Quick preview, large files  |
| 9     | Moderate | Medium | ~2 GB   | Draft processing            |
| 10    | Moderate | Good   | ~4 GB   | General use                 |
| 11    | Slow     | High   | ~8 GB   | Production quality          |
| 12    | Very slow| Very high | ~16 GB | Maximum detail            |

## Processing Pipeline

The script performs these steps automatically:

1. **[1/5] Load LAS file** into CloudComPy
2. **[2/5] Compute normals** using triangulation model with MST orientation
3. **[3/5] Convert normals** to DIP/Dip Direction scalar fields
4. **[4/5] Poisson reconstruction** with density scalar field output
5. **[5/5] Save project** as CloudCompare `.bin` file (includes color transfer)

## Output

Processed files are saved in the `Processed/` subdirectory:

```
input_directory/
├── scan1.las
├── scan2.las
└── Processed/
    ├── scan1.bin    # CloudCompare project
    └── scan2.bin
```

Each `.bin` file contains:
- **Point Cloud**: Original points with normals and DIP scalar fields
- **Mesh**: Reconstructed surface with RGB colors and density scalar field

## Filtering the Mesh in CloudCompare

The **Density** scalar field indicates reconstruction confidence:
- **High density** (warm colors) = reliable reconstruction
- **Low density** (cool colors) = sparse data, potential artifacts

### Filtering Workflow

1. Open the `.bin` file in CloudCompare
2. Select the mesh in DB Tree
3. Go to Properties → Scalar Fields → select "Density"
4. Adjust SF Display min value to visualize low-density areas
5. Edit → Scalar Fields → Filter by Value
6. Set minimum threshold and click Split/Export
7. Save the filtered mesh (File → Save)

## Troubleshooting

### "Conda not found" Error

Run the TUI from Anaconda Prompt, or ensure conda is in your PATH.

### "Failed to activate CloudComPy311" Error

The conda environment doesn't exist. Run `setup_cloudcompy.bat` first.

### "CloudComPy not found" Error

Verify CloudComPy binaries are extracted to `C:\bin\CloudComPy311`. If using a different path, edit `run_cloudcompy.bat`:

```batch
set CLOUDCOMPY_PATH=C:\your\path\to\CloudComPy311
```

### "CloudComPy import failed" Error

The environment setup may be incomplete. Try:
```batch
conda activate CloudComPy311
cd C:\bin\CloudComPy311
envCloudComPy.bat
python -c "import cloudComPy"
```

### "PoissonRecon plugin not available" Error

Ensure you downloaded the full CloudComPy package that includes plugins. The PoissonRecon plugin should be in:
```
C:\bin\CloudComPy311\CloudCompare\plugins\
```

### Processing Takes Very Long

The Poisson reconstruction step (Step 4) is computationally intensive:
- **Depth 8**: 1-5 minutes
- **Depth 11**: 10-45+ minutes for large files

The animated progress bar shows estimated progress. Reduce octree depth for faster processing.

### Mesh Has No Colors

The script automatically transfers colors if the source LAS has RGB values. If colors are missing:
- Verify the LAS file contains RGB data
- Check for "Colors transferred to mesh" in the output log

### Memory Errors

For large point clouds:
- Use lower `--octree-depth` (8 or 9)
- Close other applications to free RAM
- Process files one at a time

## File Structure

```
cloudcompare-automation/
├── README.md                   # This file
├── setup_cloudcompy.bat        # Conda environment setup script
├── run_cloudcompy.bat          # Wrapper to run with correct environment
├── process_las_files.py        # Main processing script
├── build.bat                   # Build script for the TUI
├── go.mod                      # Go module definition
├── cmd/
│   └── cloudcompare-tui/
│       └── main.go             # TUI entry point
└── internal/
    ├── tui/
    │   ├── model.go            # Bubble Tea model & animations
    │   ├── views.go            # Screen rendering
    │   └── styles.go           # Lipgloss styling
    └── processor/
        └── processor.go        # Python script integration
```

## License

MIT License

## Acknowledgments

- [CloudCompare](https://www.cloudcompare.org/) - 3D point cloud processing software
- [CloudComPy](https://github.com/CloudCompare/CloudComPy) - Python bindings for CloudCompare
- [PoissonRecon](https://github.com/mkazhdan/PoissonRecon) - Surface reconstruction algorithm
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework for Go
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal apps
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
