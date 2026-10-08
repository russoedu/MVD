import { logoWord, pixelLetters } from './pixel-logo.content'

const LETTER_WIDTH = 5
const LETTER_HEIGHT = 7
const GAP = 1
const SHADOW = 1

/** The MVD logo drawn pixel by pixel: neon green with a Barney-purple drop shadow. */
export function PixelLogo () {
  const letters = [...logoWord]
  const width = letters.length * (LETTER_WIDTH + GAP) - GAP + SHADOW
  const height = LETTER_HEIGHT + SHADOW

  const pixels = (colour: string, offset: number) => letters.flatMap((letter, index) =>
    pixelLetters[letter].flatMap((row, y) => [...row].flatMap((cell, x) => cell === '#'
      ? [
          <rect
            key={`${colour}-${index}-${x}-${y}`}
            x={index * (LETTER_WIDTH + GAP) + x + offset}
            y={y + offset}
            width={1}
            height={1}
            fill={colour}
          />,
        ]
      : [])))

  return (
    <svg
      className='pixel-logo'
      viewBox={`0 0 ${width} ${height}`}
      shapeRendering='crispEdges'
      role='img'
      aria-label='MVD'
    >
      {pixels('#a020f0', SHADOW)}
      {pixels('#55ff55', 0)}
    </svg>
  )
}
