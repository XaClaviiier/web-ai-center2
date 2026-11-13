/* Ilustrasi konseptual tahap 1, dari <template id="ig-gambar-1"> di prototipe. Bukan citra pasien. */
import type { DiagramProps } from './ImplanGigi';

export default function ImplantDiagram1({ label, nyala }: DiagramProps) {
  return (
    <svg role="img" aria-label={label} className="ig-svg" viewBox="0 0 800 600">
      <defs>
        <pattern id="ig1" width="16" height="16" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
          <rect width="16" height="16" className="ig-arsir-latar" />
          <line x1="0" y1="0" x2="0" y2="16" className="ig-arsir-garis" />
        </pattern>
        <marker id="ig1-u" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M1 1 L9 5 L1 9 z" className="ig-ujung" />
        </marker>
      </defs>
      <path d="M567.6 139.1 A228 402 0 0 1 112.4 139.1" className="ig-rahang-tepi" strokeWidth="134" />
      <path d="M567.6 139.1 A228 402 0 0 1 112.4 139.1" className="ig-rahang-isi" strokeWidth="128" />
      <g>
        <rect x="-37.2" y="-40.8" width="74.4" height="81.6" rx="18.6" transform="translate(562.9 199.8) rotate(97)" className="ig-gigi" />
        <rect x="-37.2" y="-40.8" width="74.4" height="81.6" rx="18.6" transform="translate(548.2 278.8) rotate(104.2)" className="ig-gigi" />
        <rect x="-27.6" y="-36" width="55.2" height="72" rx="13.8" transform="translate(489.9 417.9) rotate(123)" className="ig-gigi" />
        <rect x="-26.4" y="-36" width="52.8" height="72" rx="13.2" transform="translate(452.6 464.5) rotate(134.9)" className="ig-gigi" />
        <rect x="-22.8" y="-31.2" width="45.6" height="62.4" rx="11.4" transform="translate(409.2 498.1) rotate(150.7)" className="ig-gigi" />
        <rect x="-20.4" y="-31.2" width="40.8" height="62.4" rx="10.2" transform="translate(363 514.9) rotate(169.8)" className="ig-gigi" />
        <rect x="-20.4" y="-31.2" width="40.8" height="62.4" rx="10.2" transform="translate(316.6 514.9) rotate(-169.7)" className="ig-gigi" />
        <rect x="-22.8" y="-31.2" width="45.6" height="62.4" rx="11.4" transform="translate(270.5 497.9) rotate(-150.6)" className="ig-gigi" />
        <rect x="-26.4" y="-36" width="52.8" height="72" rx="13.2" transform="translate(227.1 464.2) rotate(-134.8)" className="ig-gigi" />
        <rect x="-27.6" y="-36" width="55.2" height="72" rx="13.8" transform="translate(189.8 417.5) rotate(-122.9)" className="ig-gigi" />
        <rect x="-37.2" y="-40.8" width="74.4" height="81.6" rx="18.6" transform="translate(156.9 354.6) rotate(-112.8)" className="ig-gigi" />
        <rect x="-37.2" y="-40.8" width="74.4" height="81.6" rx="18.6" transform="translate(131.6 278.2) rotate(-104.1)" className="ig-gigi" />
        <rect x="-37.2" y="-40.8" width="74.4" height="81.6" rx="18.6" transform="translate(117.1 199.2) rotate(-96.9)" className="ig-gigi" />
      </g>
      <g data-lapisan="area" className={nyala['area'] ? undefined : 'ig-mati'}>
        <rect x="-37.2" y="-40.8" width="74.4" height="81.6" rx="18.6" transform="translate(522.9 355.1) rotate(112.9)" className="ig-target" />
        <line x1="485.7" y1="355.1" x2="340" y2="210" className="ig-penghubung ig-penghubung--target" />
        <g className="ig-pil ig-pil--target">
          <rect x="203" y="187" width="274" height="46" rx="23" />
          <text x="340" y="219" textAnchor="middle">Area edentulous</text>
        </g>
      </g>
      <g data-lapisan="penanda" className={nyala['penanda'] ? undefined : 'ig-mati'}>
        <line x1="397.4" y1="409.4" x2="523.4" y2="535.2" className="ig-slice" />
        <circle cx="543.2" cy="555" r="25" className="ig-tanda" />
        <text x="543.2" y="565" textAnchor="middle" className="ig-nomor">1</text>
        <line x1="424.5" y1="375.4" x2="573.8" y2="472.3" className="ig-slice" />
        <circle cx="597.3" cy="487.6" r="25" className="ig-tanda" />
        <text x="597.3" y="497.6" textAnchor="middle" className="ig-nomor">2</text>
        <line x1="451" y1="324.8" x2="615" y2="394" className="ig-slice ig-slice--edentulous" />
        <rect x="618.8" y="382.9" width="44" height="44" rx="5" transform="rotate(45 640.8 404.9)" className="ig-tanda ig-tanda--edentulous" />
        <text x="640.8" y="414.9" textAnchor="middle" className="ig-nomor">3</text>
        <line x1="472.6" y1="259.6" x2="645.2" y2="303.3" className="ig-slice" />
        <circle cx="672.3" cy="310.2" r="25" className="ig-tanda" />
        <text x="672.3" y="320.2" textAnchor="middle" className="ig-nomor">4</text>
        <line x1="485.4" y1="190.3" x2="662.1" y2="212" className="ig-slice" />
        <circle cx="689.9" cy="215.4" r="25" className="ig-tanda" />
        <text x="689.9" y="225.4" textAnchor="middle" className="ig-nomor">5</text>
      </g>
    </svg>
  );
}
