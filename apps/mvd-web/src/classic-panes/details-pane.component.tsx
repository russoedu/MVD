import type { DetailsSnapshot } from '../workspace-snapshot'

interface DetailsPaneProps {
  details: DetailsSnapshot
}

/** What MVD says about the chosen song or source, in the theme's fixed-width font. */
export function DetailsPane ({ details }: DetailsPaneProps) {
  return (
    <section className='classic-pane classic-details' aria-label={details.heading}>
      <h2 className='classic-pane-heading'>{details.heading}</h2>
      <div className='classic-pane-body classic-log' role='log'>
        {details.lines.map((line, index) => <div key={`${index}:${line}`}>{line}</div>)}
      </div>
    </section>
  )
}
