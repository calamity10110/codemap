"""
Cityscape visualization for codemap.
Renders codebase as a city skyline where each building represents a language/file type.
"""
import os
import math
import random
from collections import defaultdict
from rich.console import Console

# What counts as "source code" (buildings) vs assets/config (excluded)
CODE_EXTENSIONS = {
    # Languages
    '.py', '.js', '.ts', '.jsx', '.tsx', '.go', '.rs', '.rb', '.java',
    '.swift', '.kt', '.scala', '.c', '.cpp', '.h', '.hpp', '.cs', '.fs',
    '.php', '.lua', '.r', '.dart', '.vue', '.svelte', '.elm', '.ex', '.exs',
    '.hs', '.ml', '.clj', '.erl', '.sh', '.bash', '.zsh', '.fish', '.ps1',
    # Markup/Style that's "code-like"
    '.html', '.css', '.scss', '.sass', '.less',
    # Config that's essential
    '.sql', '.graphql', '.proto',
}

# Files without extensions that are typically code
CODE_FILENAMES = {
    'Makefile', 'Dockerfile', 'Rakefile', 'Gemfile', 'Procfile',
    'Vagrantfile', 'Jenkinsfile', 'Fastfile',
}

# Building dimensions
BUILDING_WIDTH = 7
MAX_HEIGHT = 12
MIN_HEIGHT = 2
SKY_HEIGHT = 6


def get_file_color(ext):
    """Return a color style based on file extension."""
    ext = ext.lower()
    if ext in ['.go', '.mod', '.sum', '.dart']:
        return "cyan"
    elif ext in ['.py', '.js', '.ts', '.jsx', '.tsx', '.vue', '.svelte']:
        return "yellow"
    elif ext in ['.html', '.css', '.scss', '.sass', '.less', '.php']:
        return "magenta"
    elif ext in ['.md', '.txt', '.rst']:
        return "green"
    elif ext in ['.rb', '.erb']:
        return "red"
    elif ext in ['.sh', '.bash', 'makefile', 'dockerfile']:
        return "white"
    elif ext in ['.swift', '.kt', '.java', '.rs']:
        return "red"
    elif ext in ['.c', '.cpp', '.h', '.hpp', '.cs']:
        return "blue"
    return "white"


def get_building_char(ext):
    """Return building texture character based on file type."""
    ext = ext.lower()
    if ext in ['.go', '.mod', '.sum', '.dart']:
        return '▓'
    elif ext in ['.py', '.js', '.ts', '.jsx', '.tsx']:
        return '░'
    elif ext in ['.rb', '.erb']:
        return '▒'
    elif ext in ['.sh', 'makefile', 'dockerfile']:
        return '█'
    return '▓'


def format_size(size):
    """Format bytes to human readable."""
    for unit in ['B', 'KB', 'MB', 'GB']:
        if size < 1024:
            return f"{size:.1f}{unit}"
        size /= 1024
    return f"{size:.1f}TB"


def filter_code_files(files):
    """Filter to only source code files."""
    code_files = [
        f for f in files
        if f['ext'].lower() in CODE_EXTENSIONS
        or os.path.basename(f['path']) in CODE_FILENAMES
    ]
    return code_files if code_files else files


def aggregate_by_extension(code_files):
    """Group files by extension, summing sizes."""
    ext_groups = defaultdict(lambda: {'size': 0, 'count': 0, 'files': []})

    for f in code_files:
        ext = f['ext'].lower() if f['ext'] else os.path.basename(f['path'])
        ext_groups[ext]['size'] += f['size']
        ext_groups[ext]['count'] += 1
        ext_groups[ext]['files'].append(f)

    aggregated = [
        {'ext': ext, 'size': data['size'], 'count': data['count'], 'files': data['files']}
        for ext, data in ext_groups.items()
    ]
    return sorted(aggregated, key=lambda x: x['size'], reverse=True)


def create_buildings(sorted_files, width):
    """Create building data from aggregated files."""
    if not sorted_files:
        return []

    sizes = [f['size'] for f in sorted_files]
    min_size, max_size = min(sizes), max(sizes)
    size_range = max_size - min_size if max_size > min_size else 1

    def get_height(size):
        ratio = (size - min_size) / size_range
        ratio = math.sqrt(ratio)  # Spread out middle range
        return int(MIN_HEIGHT + ratio * (MAX_HEIGHT - MIN_HEIGHT))

    random.seed(42)
    building_data = []

    for agg in sorted_files:
        ext_label = agg['ext'] if agg['ext'].startswith('.') else agg['ext'][:7]
        building_data.append({
            'height': get_height(agg['size']),
            'char': get_building_char(agg['ext']),
            'color': get_file_color(agg['ext']),
            'ext': agg['ext'],
            'ext_label': ext_label[:5],
            'count': agg['count'],
            'size': agg['size'],
            'gap': random.choice([1, 2, 2, 3]),
        })

    # Arrange: tallest in middle, shorter on edges
    building_data.sort(key=lambda x: x['height'], reverse=True)
    arranged = []
    for i, b in enumerate(building_data):
        if i % 2 == 0:
            arranged.append(b)
        else:
            arranged.insert(0, b)

    # Limit to fit width
    total_width = sum(BUILDING_WIDTH + b['gap'] for b in arranged)
    while total_width > width - 8 and arranged:
        removed = arranged.pop(0) if len(arranged) % 2 == 0 else arranged.pop()
        total_width -= (BUILDING_WIDTH + removed['gap'])

    return arranged


def render_sky(width, scene_left, scene_right, sky_height):
    """Generate starry sky lines."""
    scene_width = scene_right - scene_left
    sky_lines = []

    for row in range(sky_height):
        line = [' '] * width
        for _ in range(scene_width // 10):
            col = random.randint(scene_left, scene_right - 1)
            if 0 <= col < width:
                line[col] = random.choice(['·', '·', '·', '✦', '*', '·'])
        sky_lines.append(line)

    # Moon
    moon_col = scene_right - 3
    if 1 < len(sky_lines) and 0 <= moon_col < width:
        sky_lines[1][moon_col] = '◐'

    return sky_lines


def render_buildings(arranged, left_margin, width):
    """Render building rows."""
    building_rows = [[' '] * width for _ in range(MAX_HEIGHT + 1)]

    col = left_margin
    for b in arranged:
        building_top = MAX_HEIGHT - b['height']

        # Rooftop cap
        if building_top > 0:
            for j in range(BUILDING_WIDTH):
                if col + j < width:
                    building_rows[building_top][col + j] = '▄'

        # Building body with extension window
        building_height = MAX_HEIGHT + 1 - building_top - 1
        center_row = building_top + 1 + building_height // 2
        ext_label = b['ext_label']

        for row in range(building_top + 1, MAX_HEIGHT + 1):
            for j in range(BUILDING_WIDTH):
                if col + j < width:
                    if row == center_row and building_height >= 3:
                        ext_start = (BUILDING_WIDTH - len(ext_label)) // 2
                        ext_end = ext_start + len(ext_label)
                        if ext_start <= j < ext_end:
                            building_rows[row][col + j] = ext_label[j - ext_start]
                        else:
                            building_rows[row][col + j] = b['char']
                    else:
                        building_rows[row][col + j] = b['char']

        col += BUILDING_WIDTH + b['gap']

    return building_rows


def colorize_output(grid, arranged, left_margin, scene_left, scene_width, width):
    """Apply colors to the grid and return styled lines."""
    output_lines = []

    # Sky
    for line in grid[:SKY_HEIGHT]:
        styled = ''
        for ch in line:
            if ch == '◐':
                styled += f'[bold yellow]{ch}[/]'
            elif ch in ['·', '✦', '*']:
                styled += f'[dim white]{ch}[/]'
            else:
                styled += ' '
        output_lines.append(styled)

    # Building positions for coloring
    col_positions = []
    col = left_margin
    for b in arranged:
        col_positions.append((col, col + BUILDING_WIDTH, b['color']))
        col += BUILDING_WIDTH + b['gap']

    # Buildings
    for line in grid[SKY_HEIGHT:]:
        styled = ''
        for char_idx, ch in enumerate(line):
            if ch == ' ':
                styled += ' '
            elif ch == '▄':
                color = 'white'
                for start, end, c in col_positions:
                    if start <= char_idx < end:
                        color = c
                        break
                styled += f'[{color}]{ch}[/]'
            elif ch == '.' or (ch.isalpha() and ch.islower()):
                styled += f'[dim white]{ch}[/]'
            elif ch.isalnum() or ch in '_-':
                styled += f'[bold white]{ch}[/]'
            else:
                color = 'white'
                for start, end, c in col_positions:
                    if start <= char_idx < end:
                        color = c
                        break
                styled += f'[{color}]{ch}[/]'
        output_lines.append(styled)

    # Ground
    ground = ' ' * max(0, scene_left) + '▀' * scene_width
    output_lines.append(f'[dim white]{ground}[/]')

    return output_lines


def render(files, project_name):
    """Main entry point: render codebase as cityscape."""
    console = Console()
    width = console.width or 80

    # Filter and aggregate
    code_files = filter_code_files(files)
    sorted_files = aggregate_by_extension(code_files)

    # Create buildings
    arranged = create_buildings(sorted_files, width)
    if not arranged:
        console.print("[dim]No source files to display[/]")
        return

    # Calculate layout
    total_width = sum(BUILDING_WIDTH + b['gap'] for b in arranged)
    left_margin = (width - total_width) // 2
    scene_padding = 4
    scene_left = max(0, left_margin - scene_padding)
    scene_right = min(width, left_margin + total_width + scene_padding)
    scene_width = scene_right - scene_left

    # Build grid
    sky = render_sky(width, scene_left, scene_right, SKY_HEIGHT)
    buildings = render_buildings(arranged, left_margin, width)
    grid = sky + buildings

    # Colorize and output
    output_lines = colorize_output(grid, arranged, left_margin, scene_left, scene_width, width)

    console.print()
    for line in output_lines:
        console.print(line)

    # Stats
    console.print()
    title = f"─── {project_name} ───"
    console.print(f"[bold white]{title:^{width}}[/]")

    code_size = sum(f['size'] for f in code_files)
    stats = f"[cyan]{len(sorted_files)}[/] languages · [cyan]{len(code_files)}[/] files · [cyan]{format_size(code_size)}[/]"
    console.print(f"{stats:^{width + 20}}")  # Extra for markup
    console.print()
