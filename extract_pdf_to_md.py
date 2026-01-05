#!/usr/bin/env python3
"""
Extract text from PDF and convert to Markdown
"""
import sys
import pdfplumber
import re

def extract_pdf_to_markdown(pdf_path, output_path):
    """Extract text from PDF and convert to basic Markdown"""

    print(f"Extracting text from: {pdf_path}")
    print(f"Output to: {output_path}")

    markdown_content = []

    with pdfplumber.open(pdf_path) as pdf:
        total_pages = len(pdf.pages)
        print(f"Total pages: {total_pages}")

        for i, page in enumerate(pdf.pages, 1):
            print(f"Processing page {i}/{total_pages}...", end='\r')

            # Extract text
            text = page.extract_text()

            if text:
                # Add page separator
                if i > 1:
                    markdown_content.append(f"\n\n---\n\n")

                # Add page number as comment
                markdown_content.append(f"<!-- Page {i} -->\n\n")

                # Process text line by line
                lines = text.split('\n')
                processed_lines = []

                for line in lines:
                    line = line.strip()
                    if not line:
                        continue

                    # Detect potential headers (all caps, short lines)
                    if len(line) < 100 and line.isupper() and len(line.split()) <= 10:
                        # Check if it looks like a header (not just page numbers or single words)
                        if len(line) > 3 and not line.isdigit():
                            processed_lines.append(f"\n## {line.title()}\n")
                        else:
                            processed_lines.append(line)
                    else:
                        processed_lines.append(line)

                markdown_content.append('\n'.join(processed_lines))

        print(f"\nProcessing complete. Pages processed: {total_pages}")

    # Write to file
    full_content = ''.join(markdown_content)

    # Add header
    header = f"# Règles de Base - Basic Fantasy RPG\n\n"
    header += f"*Extrait automatiquement du PDF source*\n\n"

    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(header + full_content)

    print(f"\n✓ Markdown file created: {output_path}")
    print(f"  Size: {len(full_content)} characters")

    return output_path

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python3 extract_pdf_to_md.py <input.pdf> <output.md>")
        sys.exit(1)

    pdf_path = sys.argv[1]
    output_path = sys.argv[2]

    extract_pdf_to_markdown(pdf_path, output_path)
