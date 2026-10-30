/** The MVD letters as 5x7 bitmaps, one string per row, "#" for a lit pixel. */
export const pixelLetters: Record<string, string[]> = {
  M: ['#...#', '##.##', '#.#.#', '#.#.#', '#...#', '#...#', '#...#'],
  V: ['#...#', '#...#', '#...#', '#...#', '#...#', '.#.#.', '..#..'],
  D: ['####.', '#...#', '#...#', '#...#', '#...#', '#...#', '####.'],
}

/** The word the pixel logo spells. Each of its letters has a bitmap in pixelLetters. */
export const logoWord = 'MVD'
