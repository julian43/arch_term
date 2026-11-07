#!/usr/bin/env python3
"""
ArchTerm Python Parser
Converts PDF/MD files to structured JSON for the ArchTerm application.
"""

import argparse
import json
import os
import sys
from pathlib import Path


def parse_pdf(filepath: str) -> dict:
    """Parse a PDF file and return structured data."""
    try:
        import fitz  # PyMuPDF
    except ImportError:
        print("Error: PyMuPDF is required. Install with: pip install pymupdf", file=sys.stderr)
        sys.exit(1)

    result = {
        "filename": os.path.basename(filepath),
        "type": "pdf",
        "pages": [],
        "metadata": {}
    }

    try:
        doc = fitz.open(filepath)
        result["metadata"]["page_count"] = doc.page_count
        result["metadata"]["title"] = doc.metadata.get("title", "Unknown")
        result["metadata"]["author"] = doc.metadata.get("author", "Unknown")

        for page_num in range(min(10, doc.page_count)):  # Limit to first 10 pages
            page = doc[page_num]
            text = page.get_text()
            result["pages"].append({
                "page": page_num + 1,
                "text": text
            })
        doc.close()
    except Exception as e:
        result["error"] = str(e)

    return result


def parse_markdown(filepath: str) -> dict:
    """Parse a Markdown file and return structured data."""
    try:
        import markdown
    except ImportError:
        print("Error: Python-Markdown is required. Install with: pip install markdown", file=sys.stderr)
        sys.exit(1)

    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    html = markdown.markdown(content, extensions=['tables', 'fenced-code'])

    return {
        "filename": os.path.basename(filepath),
        "type": "markdown",
        "content": content,
        "html": html,
        "metadata": {
            "lines": content.count('\n') + 1,
            "words": len(content.split()),
            "characters": len(content)
        }
    }


def parse_text(filepath: str) -> dict:
    """Parse a text file and return structured data."""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    return {
        "filename": os.path.basename(filepath),
        "type": "text",
        "content": content,
        "metadata": {
            "lines": content.count('\n') + 1,
            "words": len(content.split()),
            "characters": len(content)
        }
    }


def parse_json_file(filepath: str) -> dict:
    """Parse a JSON file and return structured data."""
    with open(filepath, 'r', encoding='utf-8') as f:
        data = json.load(f)

    return {
        "filename": os.path.basename(filepath),
        "type": "json",
        "content": data,
        "metadata": {
            "keys": list(data.keys()) if isinstance(data, dict) else []
        }
    }


def parse_file(filepath: str) -> dict:
    """Parse a file based on its extension."""
    ext = os.path.splitext(filepath)[1].lower()
    filename = os.path.basename(filepath)

    # Update path to include full path
    filepath = os.path.abspath(filepath)

    if ext == '.pdf':
        return parse_pdf(filepath)
    elif ext in ('.md', '.markdown'):
        return parse_markdown(filepath)
    elif ext == '.json':
        return parse_json_file(filepath)
    else:
        return parse_text(filepath)


def main():
    parser = argparse.ArgumentParser(description='Parse files for ArchTerm')
    parser.add_argument('--input', '-i', required=True, help='Input file path')
    parser.add_argument('--output', '-o', default='-', help='Output file path (default: stdout)')
    parser.add_argument('--format', '-f', choices=['json', 'text'], default='json', help='Output format')
    args = parser.parse_args()

    # Parse the file
    result = parse_file(args.input)

    # Output the result
    if args.output == '-':
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        with open(args.output, 'w', encoding='utf-8') as f:
            json.dump(result, f, indent=2, ensure_ascii=False)
        print(f"Output written to {args.output}")


if __name__ == '__main__':
    main()
