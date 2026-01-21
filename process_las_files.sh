#!/bin/bash

# ============================================================================
# CloudCompare Batch Processing Script for LAS Files
# ============================================================================
# This script processes LAS point cloud files through CloudCompare CLI to:
# 1. Import LAS files
# 2. Compute normals (Triangulation + MST orientation knn=6)
# 3. Convert normals to Dip/Dip Direction
# 4. Perform Poisson Surface Reconstruction
# 5. Save the project as .bin file
# ============================================================================

# Configuration
# ----------------------------------------------------------------------------
# Path to CloudCompare executable (adjust based on your installation)
# Linux typical paths:
#   - /usr/bin/CloudCompare
#   - /usr/bin/cloudcompare
#   - /opt/CloudCompare/CloudCompare
#   - ~/CloudCompare/CloudCompare
# macOS typical path:
#   - /Applications/CloudCompare.app/Contents/MacOS/CloudCompare

CLOUDCOMPARE_PATH="${CLOUDCOMPARE_PATH:-CloudCompare}"

# Input directory containing LAS files
INPUT_DIR="${1:-.}"

# Output directory for processed files (relative to input)
OUTPUT_DIR="Processed"

# Poisson Reconstruction Parameters
OCTREE_DEPTH=11
SAMPLES_PER_NODE=1.5
POINT_WEIGHT=2.0
THREADS=16
BOUNDARY="NEUMANN"

# Normal Computation Parameters
KNN=6

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

check_cloudcompare() {
    if ! command -v "$CLOUDCOMPARE_PATH" &> /dev/null; then
        # Try common alternative paths
        local alternatives=(
            "cloudcompare"
            "CloudCompare"
            "/usr/bin/CloudCompare"
            "/usr/bin/cloudcompare"
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
        print_error "Example: export CLOUDCOMPARE_PATH=/path/to/CloudCompare"
        exit 1
    fi
    print_info "Using CloudCompare: $CLOUDCOMPARE_PATH"
}

# ============================================================================
# Main Processing Function
# ============================================================================

process_las_file() {
    local input_file="$1"
    local filename=$(basename "$input_file" .las)
    filename=$(basename "$filename" .LAS)  # Handle uppercase extension
    local output_file="${OUTPUT_DIR}/${filename}.bin"

    print_header "Processing: $filename"
    print_info "Input:  $input_file"
    print_info "Output: $output_file"

    # CloudCompare CLI command
    # Note: CloudCompare CLI processes commands in sequence
    #
    # Command breakdown:
    # -SILENT              : Suppress GUI dialogs
    # -O                   : Open/load file
    # -COMPUTE_NORMALS     : Compute normals for the point cloud
    # -NORMALS_TO_DIP      : Convert normals to Dip/Dip Direction
    # -POISSON             : Poisson Surface Reconstruction
    # -C_EXPORT_FMT BIN    : Set cloud export format to BIN
    # -M_EXPORT_FMT BIN    : Set mesh export format to BIN
    # -SAVE_CLOUDS         : Save point clouds
    # -SAVE_MESHES         : Save meshes

    # Build the command
    # Note: Some parameters may need adjustment based on CloudCompare version
    local cmd=(
        "$CLOUDCOMPARE_PATH"
        -SILENT
        -AUTO_SAVE OFF
        -O "$input_file"

        # Compute Normals
        # LOCAL_MODEL: TRI (Triangulation), QUADRIC, or PLANE
        # ORIENT: PLUS_ZERO, MINUS_ZERO, PLUS_BARYCENTER, MINUS_BARYCENTER,
        #         PLUS_X, MINUS_X, PLUS_Y, MINUS_Y, PLUS_Z, MINUS_Z,
        #         PREVIOUS, or MST with neighbor count
        -COMPUTE_NORMALS

        # Convert Normals to Dip/Dip Direction
        -NORMALS_TO_DIP

        # Poisson Surface Reconstruction
        # Parameters: OCTREE_DEPTH SAMPLES_PER_NODE BOUNDARY
        # Boundary types: FREE, DIRICHLET, NEUMANN
        -POISSON "$OCTREE_DEPTH" "$SAMPLES_PER_NODE" "$BOUNDARY"

        # Set export formats
        -C_EXPORT_FMT BIN
        -M_EXPORT_FMT BIN

        # Save outputs to the processed directory
        -SAVE_CLOUDS FILE "$output_file"
    )

    print_info "Executing CloudCompare CLI..."
    print_info "Command: ${cmd[*]}"
    echo ""

    # Execute the command
    if "${cmd[@]}"; then
        print_success "Successfully processed: $filename"
        return 0
    else
        print_error "Failed to process: $filename"
        return 1
    fi
}

# Alternative processing using a more compatible approach
# Some CloudCompare versions have different CLI syntax
process_las_file_alternative() {
    local input_file="$1"
    local filename=$(basename "$input_file" .las)
    filename=$(basename "$filename" .LAS)
    local output_file="${OUTPUT_DIR}/${filename}.bin"

    print_header "Processing: $filename (Alternative Method)"
    print_info "Input:  $input_file"
    print_info "Output: $output_file"

    # This is a more verbose command that may work better with some versions
    "$CLOUDCOMPARE_PATH" -SILENT -AUTO_SAVE OFF \
        -O "$input_file" \
        -COMPUTE_NORMALS \
        -NORMALS_TO_DIP \
        -POISSON "$OCTREE_DEPTH" \
        -C_EXPORT_FMT BIN \
        -SAVE_CLOUDS FILE "$output_file"

    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        print_success "Successfully processed: $filename"
    else
        print_error "Failed to process: $filename (exit code: $exit_code)"
    fi

    return $exit_code
}

# ============================================================================
# Main Script
# ============================================================================

main() {
    print_header "CloudCompare Batch Processing Script"

    # Check if CloudCompare is available
    check_cloudcompare

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

    # Find all LAS files
    shopt -s nullglob nocaseglob
    las_files=(*.las *.LAS)
    shopt -u nullglob nocaseglob

    if [ ${#las_files[@]} -eq 0 ]; then
        print_error "No LAS files found in: $INPUT_DIR"
        exit 1
    fi

    print_info "Found ${#las_files[@]} LAS file(s) to process"
    echo ""

    # Process each file
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

    return $fail_count
}

# ============================================================================
# Script Entry Point
# ============================================================================

# Show usage if -h or --help is passed
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    echo "Usage: $0 [INPUT_DIRECTORY]"
    echo ""
    echo "Process LAS point cloud files with CloudCompare."
    echo ""
    echo "Arguments:"
    echo "  INPUT_DIRECTORY   Directory containing LAS files (default: current directory)"
    echo ""
    echo "Environment Variables:"
    echo "  CLOUDCOMPARE_PATH   Path to CloudCompare executable"
    echo ""
    echo "Output:"
    echo "  Processed files will be saved to INPUT_DIRECTORY/Processed/"
    echo ""
    echo "Processing Steps:"
    echo "  1. Import LAS file"
    echo "  2. Compute Normals (Triangulation, MST orientation knn=6)"
    echo "  3. Convert Normals to Dip/Dip Direction"
    echo "  4. Poisson Surface Reconstruction (depth=$OCTREE_DEPTH)"
    echo "  5. Save as .bin project file"
    exit 0
fi

# Run main function
main "$@"
