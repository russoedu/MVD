import { useState } from 'react'

/** How many times the tagline scrolls by before it stays still. */
const passes = 3

interface TickerProps {
  text: string
}

/**
 * The scrolling tagline of the header. It pauses while the pointer is over the header, stops
 * after a few passes (it can be read still, and starts again on the next start of the app),
 * and is read out once as plain text rather than as moving text.
 */
export function Ticker ({ text }: TickerProps) {
  const [seen, setSeen] = useState(0)
  const settled = seen >= passes

  return (
    <div className='page-header-ticker' data-settled={settled}>
      <span className='visually-hidden'>{text}</span>
      <span aria-hidden='true' onAnimationIteration={() => setSeen(count => count + 1)}>
        {text}
      </span>
    </div>
  )
}
