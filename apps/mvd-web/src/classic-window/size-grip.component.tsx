import { useRef } from 'react'
import { resizeWindow, windowSize } from '../window-actions'

/** The size the window had and where the pointer was when the drag began, in screen points. */
interface Drag {
  width:  number
  height: number
  x:      number
  y:      number
}

/** The smallest the window may be made from the grip, as the program allows it (see appwindow). */
const smallest = { width: 640, height: 400 }

/**
 * The corner the window is made bigger or smaller by: the size box of a classic Mac, the size grip
 * of Windows 95 onwards, the ridged corner of Mac OS X. Dragging it resizes MVD's window.
 */
export function SizeGrip () {
  const dragRef = useRef<Drag | null>(null)

  return (
    <span
      className='classic-size-grip'
      aria-hidden='true'
      onPointerDown={event => {
        if (event.button !== 0) return
        const target = event.currentTarget
        const { screenX: x, screenY: y, pointerId } = event
        target.setPointerCapture(pointerId)
        void windowSize().then(size => {
          if (size) dragRef.current = { ...size, x, y }
        })
      }}
      onPointerMove={event => {
        const drag = dragRef.current
        if (!drag) return
        resizeWindow(
          Math.max(drag.width + event.screenX - drag.x, smallest.width),
          Math.max(drag.height + event.screenY - drag.y, smallest.height),
        )
      }}
      onPointerUp={() => { dragRef.current = null }}
      onPointerCancel={() => { dragRef.current = null }}
    />
  )
}
