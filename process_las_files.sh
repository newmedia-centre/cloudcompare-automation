#!/bin/bash

# ============================================================================
# CloudCompare + PoissonRecon Batch Processing Script for LAS Files
# ============================================================================
# This script processes LAS point cloud files through:
# 1. CloudCompare CLI - Import LAS, compute normals, convert to DIP, export PLY
# 2. PoissonRecon - Poisson Surface Reconstruction with density scalar field
# 3. CloudCompare CLI - Import point cloud and mesh into project for manual
#                       filtering by density SF value
#
# The density scalar field from PoissonRecon allows you to filter out
# low-confidence areas of the reconstruction in CloudCompare by adjusting
# the SF display min value and exporting the filtered mesh.
#
# Prerequisites:
# - CloudCompare (with CLI support)
# - PoissonRecon from https://github.com/mkazhdan/PoissonRecon
#
# Poisson Surface Reconstruction Parameters:
# - Octree depth = 11
# - Boundary = Neumann
# - Samples per node = 1.5
# - Point weight = 2.0
# - Threads = 16
# - Output density as SF = Yes (for manual filtering in CloudCompare)
# ============================================================================

set -e  # Exit on error

# Configuration
# ----------------------------------------------------------------------------
# Path to CloudCompare executable
CLOUDCOMPARE_PATH="${CLOUDCOMPARE_PATH:-CloudCompare}"

# Path to PoissonRecon executable
POISSONRECON_PATH="${POISSONRECON_PATH:-PoissonRecon}"

# Input directory containing LAS files
INPUT_DIR="${1:-.}"

# Output directory for processed files (relative to input)
OUTPUT_DIR="Processed"

# Temporary directory for intermediate files
TEMP_DIR=""

# Normal Computation Parameters
OCTREE_RADIUS="AUTO"
KNN=6

# Poisson Reconstruction Parameters
OCTREE_DEPTH=11
SAMPLES_PER_NODE=1.5
POINT_WEIGHT=2.0
THREADS=16
BOUNDARY_TYPE=3  # 1=Free, 2=Dirichlet, 3=Neumann

# ============================================================================
# Helper Functions
# ============================================================================

print_header() {
    echo ""
    echo "============================================================================"
    echo " $1"
    echo "============================================================================"
}

print_info() {
    echo "[INFO] $1"
}

print_error() {
    echo "[ERROR] $1" >&2
}

print_success() {
    echo "[SUCCESS] $1"
}

print_warning() {
    echo "[WARNING] $1"
}

cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        print_info "Cleaning up temporary files..."
        rm -rf "$TEMP_DIR"
    fi
}

trap cleanup EXIT

check_cloudcompare() {
    if ! command -v "$CLOUDCOMPARE_PATH" &> /dev/null; then
        local alternatives=(
            "cloudcompare"
            "CloudCompare"
            "/usr/bin/CloudCompare"
            "/usr/bin/cloudcompare"
            "/usr/local/bin/CloudCompare"
            "/opt/CloudCompare/CloudCompare"
            "/snap/bin/cloudcompare"
            "/Applications/CloudCompare.app/Contents/MacOS/CloudCompare"
        )

        for alt in "${alternatives[@]}"; do
            if command -v "$alt" &> /dev/null || [ -x "$alt" ]; then
                CLOUDCOMPARE_PATH="$alt"
                print_info "Found CloudCompare at: $CLOUDCOMPARE_PATH"
                return 0
            fi
        done

        print_error "CloudCompare not found!"
        print_error "Please set CLOUDCOMPARE_PATH environment variable or install CloudCompare."
        exit 1
    fi
    print_info "Using CloudCompare: $CLOUDCOMPARE_PATH"
}

check_poissonrecon() {
    if ! command -v "$POISSONRECON_PATH" &> /dev/null; then
        local alternatives=(
            "PoissonRecon"
            "poissonrecon"
            "/usr/local/bin/PoissonRecon"
            "/usr/bin/PoissonRecon"
            "/opt/PoissonRecon/PoissonRecon"
        )

        for alt in "${alternatives[@]}"; do
            if command -v "$alt" &> /dev/null || [ -x "$alt" ]; then
                POISSONRECON_PATH="$alt"
                print_info "Found PoissonRecon at: $POISSONRECON_PATH"
                return 0
            fi
        done

        print_error "PoissonRecon not found!"
        print_error "Please install PoissonRecon from https://github.com/mkazhdan/PoissonRecon"
        print_error "Or set POISSONRECON_PATH environment variable."
        exit 1
    fi
    print_info "Using PoissonRecon: $POISSONRECON_PATH"
}

# ============================================================================
# Processing Functions
# ============================================================================

process_las_file() {
    local input_file="$1"
    local filename=$(basename "$input_file" .las)
    filename=$(basename "$filename" .LAS)

    local ply_with_normals="${TEMP_DIR}/${filename}_normals.ply"
    local mesh_ply="${TEMP_DIR}/${filename}_mesh.ply"
    local output_bin="${OUTPUT_DIR}/${filename}.bin"

    print_header "Processing: $filename"
    print_info "Input:  $input_file"
    print_info "Output: $output_bin"

    # -------------------------------------------------------------------------
    # Step 1: CloudCompare - Load LAS, compute normals, convert to DIP, export PLY
    # -------------------------------------------------------------------------
    print_info ""
    print_info "Step 1: Computing normals and converting to DIP with CloudCompare..."
    print_info "  - Octree Radius: $OCTREE_RADIUS"
    print_info "  - MST Orientation KNN: $KNN"

    "$CLOUDCOMPARE_PATH" -SILENT -AUTO_SAVE OFF \
        -O "$input_file" \
        -OCTREE_NORMALS "$OCTREE_RADIUS" \
        -ORIENT_NORMS_MST "$KNN" \
        -NORMALS_TO_DIP \
        -C_EXPORT_FMT PLY -PLY_EXPORT_FMT ASCII \
        -SAVE_CLOUDS FILE "$ply_with_normals"

    if [ ! -f "$ply_with_normals" ]; then
        print_error "Failed to generate PLY with normals: $ply_with_normals"
        return 1
    fi
    print_success "Generated PLY with normals"

    # -------------------------------------------------------------------------
    # Step 2: PoissonRecon - Poisson Surface Reconstruction with density SF
    # -------------------------------------------------------------------------
    print_info ""
    print_info "Step 2: Running Poisson Surface Reconstruction..."
    print_info "  - Octree depth: $OCTREE_DEPTH"
    print_info "  - Samples per node: $SAMPLES_PER_NODE"
    print_info "  - Point weight: $POINT_WEIGHT"
    print_info "  - Boundary type: $BOUNDARY_TYPE (Neumann)"
    print_info "  - Threads: $THREADS"
    print_info "  - Output density as SF: YES (for filtering in CloudCompare)"

    "$POISSONRECON_PATH" \
        --in "$ply_with_normals" \
        --out "$mesh_ply" \
        --depth "$OCTREE_DEPTH" \
        --samplesPerNode "$SAMPLES_PER_NODE" \
        --pointWeight "$POINT_WEIGHT" \
        --bType "$BOUNDARY_TYPE" \
        --threads "$THREADS" \
        --density \
        --colors

    if [ ! -f "$mesh_ply" ]; then
        print_error "Failed to generate mesh: $mesh_ply"
        return 1
    fi
    print_success "Generated mesh with density scalar field"

    # -------------------------------------------------------------------------
    # Step 3: CloudCompare - Import point cloud and mesh, save as .bin project
    # -------------------------------------------------------------------------
    print_info ""
    print_info "Step 3: Creating CloudCompare project (.bin) for manual filtering..."
    print_info "  The mesh contains a 'Density' scalar field from PoissonRecon."
    print_info "  In CloudCompare, you can:"
    print_info "    1. Select the mesh"
    print_info "    2. Display the Density SF (Edit > Scalar Fields > Set Active SF)"
    print_info "    3. Adjust SF display range to visualize low-density areas"
    print_info "    4. Filter by SF value (Edit > Scalar Fields > Filter by Value)"
    print_info "    5. Export the filtered mesh"

    "$CLOUDCOMPARE_PATH" -SILENT -AUTO_SAVE OFF \
        -O "$ply_with_normals" \
        -O "$mesh_ply" \
        -SAVE_CLOUDS ALL FILE "$output_bin"

    if [ -f "$output_bin" ]; then
        print_success "Created CloudCompare project: $output_bin"
        print_info ""
        print_info "Next steps in CloudCompare GUI:"
        print_info "  1. Open $output_bin"
        print_info "  2. Select the mesh in DB Tree"
        print_info "  3. Properties panel > SF Display > adjust 'displayed' min value"
        print_info "  4. Use 'Edit > Scalar Fields > Filter by Value' to remove low-density vertices"
        print_info "  5. Export filtered mesh via 'File > Save'"
        return 0
    else
        print_error "Failed to create project: $output_bin"
        return 1
    fi
}

# ============================================================================
# Main Script
# ============================================================================

main() {
    print_header "CloudCompare + PoissonRecon Batch Processing Script"

    # Check dependencies
    check_cloudcompare
    check_poissonrecon

    # Validate input directory
    if [ ! -d "$INPUT_DIR" ]; then
        print_error "Input directory does not exist: $INPUT_DIR"
        exit 1
    fi

    # Change to input directory
    cd "$INPUT_DIR" || exit 1
    print_info "Working directory: $(pwd)"

    # Create output directory
    mkdir -p "$OUTPUT_DIR"
    print_info "Output directory: $OUTPUT_DIR"

    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    print_info "Temporary directory: $TEMP_DIR"

    # Find all LAS files
    shopt -s nullglob nocaseglob
    las_files=(*.las *.LAS)
    shopt -u nullglob nocaseglob

    if [ ${#las_files[@]} -eq 0 ]; then
        print_error "No LAS files found in: $INPUT_DIR"
        exit 1
    fi

    print_info "Found ${#las_files[@]} LAS file(s) to process"
    print_info ""
    print_info "Processing Parameters:"
    print_info "  Normal Computation:"
    print_info "    - Octree Radius: $OCTREE_RADIUS"
    print_info "    - MST Orientation KNN: $KNN"
    print_info "  Poisson Reconstruction:"
    print_info "    - Octree Depth: $OCTREE_DEPTH"
    print_info "    - Samples per Node: $SAMPLES_PER_NODE"
    print_info "    - Point Weight: $POINT_WEIGHT"
    print_info "    - Boundary: Neumann"
    print_info "    - Threads: $THREADS"
    print_info "    - Output Density as SF: YES"
    print_info "    - Interpolate Colors: YES"
    echo ""

    # Process files
    local success_count=0
    local fail_count=0

    for las_file in "${las_files[@]}"; do
        if process_las_file "$las_file"; then
            ((success_count++))
        else
            ((fail_count++))
        fi
        echo ""
    done

    # Summary
    print_header "Processing Complete"
    print_info "Successfully processed: $success_count file(s)"
    if [ $fail_count -gt 0 ]; then
        print_error "Failed to process: $fail_count file(s)"
    fi

    print_info ""
    print_info "Output files are in: $(pwd)/$OUTPUT_DIR"
    print_info "  - [filename].bin : CloudCompare project with point cloud and mesh"
    print_info ""
    print_info "The mesh includes a 'Density' scalar field from PoissonRecon."
    print_info "Open the .bin file in CloudCompare to filter out low-density areas:"
    print_info "  1. Select mesh > Properties > SF Display > adjust min value"
    print_info "  2. Edit > Scalar Fields > Filter by Value"
    print_info "  3. Export the filtered mesh"

    return $fail_count
}

# ============================================================================
# Script Entry Point
# ============================================================================

if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    echo "Usage: $0 [INPUT_DIRECTORY]"
    echo ""
    echo "Process LAS point cloud files with CloudCompare and PoissonRecon."
    echo "The output mesh includes a density scalar field for manual filtering."
    echo ""
    echo "Arguments:"
    echo "  INPUT_DIRECTORY   Directory containing LAS files (default: current directory)"
    echo ""
    echo "Environment Variables:"
    echo "  CLOUDCOMPARE_PATH   Path to CloudCompare executable"
    echo "  POISSONRECON_PATH   Path to PoissonRecon executable"
    echo ""
    echo "Output:"
    echo "  Processed files will be saved to INPUT_DIRECTORY/Processed/"
    echo "  Each .bin file contains the point cloud and mesh with density SF."
    echo ""
    echo "Processing Pipeline:"
    echo "  1. Import LAS file (CloudCompare)"
    echo "  2. Compute Normals using Octree (radius=AUTO)"
    echo "  3. Orient Normals using MST (knn=6)"
    echo "  4. Convert Normals to Dip/Dip Direction"
    echo "  5. Export as PLY with normals"
    echo "  6. Poisson Surface Reconstruction (PoissonRecon)"
    echo "     - Octree depth: 11"
    echo "     - Samples per node: 1.5"
    echo "     - Point weight: 2.0"
    echo "     - Boundary: Neumann"
    echo "     - Output density as SF: YES"
    echo "     - Interpolate colors: YES"
    echo "     - Threads: 16"
    echo "  7. Save CloudCompare project (.bin) with point cloud and mesh"
    echo ""
    echo "Manual Filtering in CloudCompare:"
    echo "  1. Open the .bin project file"
    echo "  2. Select the mesh in the DB Tree"
    echo "  3. In Properties panel, adjust SF Display min value"
    echo "  4. Use 'Edit > Scalar Fields > Filter by Value' to remove vertices"
    echo "  5. Export the filtered mesh"
    echo ""
    echo "Prerequisites:"
    echo "  - CloudCompare: https://www.cloudcompare.org/"
    echo "  - PoissonRecon: https://github.com/mkazhdan/PoissonRecon"
    exit 0
fi

main "$@"
