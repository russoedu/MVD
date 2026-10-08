import { useEffect, useRef } from 'react'
import { aboutFooter, aboutParagraphs, aboutTitle } from './about-this-app.content'
import './about-this-app.style.css'

interface AboutThisAppProps {
  open: boolean
  onClose: () => void
}

/** The "About this app" window: what MVD is for, in a Windows 95 style dialog. */
export function AboutThisApp ({ open, onClose }: AboutThisAppProps) {
  const dialog = useRef<HTMLDialogElement>(null)

  useEffect(() => {
    const element = dialog.current
    if (!element) return
    if (open && !element.open) element.showModal()
    if (!open && element.open) element.close()
  }, [open])

  return (
    <dialog ref={dialog} className='about-this-app' onClose={onClose} aria-labelledby='about-title'>
      <div className='about-title-bar'>
        <span id='about-title'>{aboutTitle}</span>
        <button type='button' className='about-close' aria-label='Close' onClick={onClose}>x</button>
      </div>
      <div className='about-body'>
        {aboutParagraphs.map(text => <p key={text}>{text}</p>)}
        <p className='about-footer'>{aboutFooter}</p>
        <button type='button' className='about-ok' onClick={onClose}>OK</button>
      </div>
    </dialog>
  )
}
