/* Ilustrasi konseptual tahap 3, dari <template id="ig-gambar-3"> di prototipe. Bukan citra pasien. */
import type { DiagramProps } from './ImplanGigi';

export default function ImplantDiagram3({ label, nyala }: DiagramProps) {
  return (
    <svg role="img" aria-label={label} className="ig-svg" viewBox="0 0 800 600">
      <defs>
        <pattern id="ig3" width="16" height="16" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
          <rect width="16" height="16" className="ig-arsir-latar" />
          <line x1="0" y1="0" x2="0" y2="16" className="ig-arsir-garis" />
        </pattern>
        <marker id="ig3-u" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M1 1 L9 5 L1 9 z" className="ig-ujung" />
        </marker>
      </defs>
      <path d="M330 152 C376.2 152 401.4 173.6 414 208.16 C426.6 281.6 470 333.44 470 411.2 C470 506.23999999999995 407 584 330 584 C253 584 190 506.23999999999995 190 411.2 C190 333.44 233.4 281.6 246 208.16 C258.6 173.6 283.8 152 330 152 Z" className="ig-tulang" />
      <clipPath id="ig3-k">
        <rect x="0" y="0" width="800" height="382" />
      </clipPath>
      <g data-lapisan="anatomi" className={nyala['anatomi'] ? undefined : 'ig-mati'}>
        <path d="M330 152 C376.2 152 401.4 173.6 414 208.16 C426.6 281.6 470 333.44 470 411.2 C470 506.23999999999995 407 584 330 584 C253 584 190 506.23999999999995 190 411.2 C190 333.44 233.4 281.6 246 208.16 C258.6 173.6 283.8 152 330 152 Z" fill="url(#ig3)" clipPath="url(#ig3-k)" className="ig-ab-tepi" />
        <circle cx="330" cy="482" r="32" className="ig-kanal" />
        <line x1="305" y1="482" x2="355" y2="482" className="ig-kanal-garis" />
        <line x1="442" y1="300" x2="580" y2="300" className="ig-penghubung" />
        <g className="ig-pil ig-pil--ab">
          <rect x="593" y="277" width="66" height="46" rx="23" />
          <text x="626" y="309" textAnchor="middle">AB</text>
        </g>
        <line x1="362" y1="482" x2="580" y2="482" className="ig-penghubung" />
        <g className="ig-pil ig-pil--mc">
          <rect x="593" y="459" width="66" height="46" rx="23" />
          <text x="626" y="491" textAnchor="middle">MC</text>
        </g>
      </g>
      <rect x="296" y="132" width="68" height="216" rx="10" className="ig-implan" />
      <path d="M296 168 H364 M296 196 H364 M296 224 H364 M296 252 H364 M296 280 H364 M296 308 H364" className="ig-implan-ulir" />
      <line x1="364" y1="170" x2="580" y2="170" className="ig-penghubung" />
      <g className="ig-pil ig-pil--implan">
        <rect x="573" y="147" width="130" height="46" rx="23" />
        <text x="638" y="179" textAnchor="middle">Implan</text>
      </g>
      <g data-lapisan="ukur" className={nyala['ukur'] ? undefined : 'ig-mati'}>
        <path d="M296 118 V70 M364 118 V70" className="ig-bantu" />
        <path d="M296 82 H364" className="ig-ukur" markerStart="url(#ig3-u)" markerEnd="url(#ig3-u)" />
        <g className="ig-pil ig-pil--ukur">
          <rect x="249" y="15" width="162" height="46" rx="23" />
          <text x="330" y="47" textAnchor="middle">Diameter</text>
        </g>
        <path d="M290 132 H150 M290 348 H150" className="ig-bantu" />
        <path d="M168 132 V348" className="ig-ukur" markerStart="url(#ig3-u)" markerEnd="url(#ig3-u)" />
        <g className="ig-pil ig-pil--ukur">
          <rect x="27" y="217" width="130" height="46" rx="23" />
          <text x="92" y="249" textAnchor="middle">Tinggi</text>
        </g>
      </g>
    </svg>
  );
}
