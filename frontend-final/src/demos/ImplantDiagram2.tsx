/* Ilustrasi konseptual tahap 2, dari <template id="ig-gambar-2"> di prototipe. Bukan citra pasien. */
import type { DiagramProps } from './ImplanGigi';

export default function ImplantDiagram2({ label, nyala, pilihan }: DiagramProps) {
  return (
    <svg role="img" aria-label={label} className="ig-svg" viewBox="0 0 800 600">
      <defs>
        <pattern id="ig2" width="16" height="16" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
          <rect width="16" height="16" className="ig-arsir-latar" />
          <line x1="0" y1="0" x2="0" y2="16" className="ig-arsir-garis" />
        </pattern>
        <marker id="ig2-u" viewBox="0 0 10 10" refX="5" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
          <path d="M1 1 L9 5 L1 9 z" className="ig-ujung" />
        </marker>
      </defs>
      <g data-slice="1" className={pilihan['slice'] === '1' ? undefined : 'ig-mati'}>
        <path d="M250 130 C283 130 301 152.5 310 188.5 C319 265 360 319 360 400 C360 499 310.5 580 250 580 C189.5 580 140 499 140 400 C140 319 181 265 190 188.5 C199 152.5 217 130 250 130 Z" className="ig-tulang" />
        <clipPath id="ig2-k0">
          <rect x="0" y="0" width="800" height="345" />
        </clipPath>
        <g data-lapisan="ab" className={nyala['ab'] ? undefined : 'ig-mati'}>
          <path d="M250 130 C283 130 301 152.5 310 188.5 C319 265 360 319 360 400 C360 499 310.5 580 250 580 C189.5 580 140 499 140 400 C140 319 181 265 190 188.5 C199 152.5 217 130 250 130 Z" fill="url(#ig2)" clipPath="url(#ig2-k0)" className="ig-ab-tepi" />
          <line x1="336" y1="265" x2="438" y2="300" className="ig-penghubung" />
          <g className="ig-pil ig-pil--ab">
            <rect x="449" y="277" width="66" height="46" rx="23" />
            <text x="482" y="309" textAnchor="middle">AB</text>
          </g>
        </g>
        <path d="M210 64 Q250 38 290 64 L286 136 L266.8 300 Q250 316 233.2 300 L214 136 Z" className="ig-gigi" />
        <g data-lapisan="mc" className={nyala['mc'] ? undefined : 'ig-mati'}>
          <circle cx="258" cy="446" r="25" className="ig-kanal" />
          <line x1="240" y1="446" x2="276" y2="446" className="ig-kanal-garis" />
          <line x1="283" y1="446" x2="438" y2="400" className="ig-penghubung" />
          <g className="ig-pil ig-pil--mc">
            <rect x="449" y="377" width="66" height="46" rx="23" />
            <text x="482" y="409" textAnchor="middle">MC</text>
          </g>
        </g>
      </g>
      <g data-slice="2" className={pilihan['slice'] === '2' ? undefined : 'ig-mati'}>
        <path d="M250 126 C285.2 126 304.4 148.7 314 185.02 C323.6 262.2 364 316.68 364 398.4 C364 498.28 312.7 580 250 580 C187.3 580 136 498.28 136 398.4 C136 316.68 176.4 262.2 186 185.02 C195.6 148.7 214.8 126 250 126 Z" className="ig-tulang" />
        <clipPath id="ig2-k1">
          <rect x="0" y="0" width="800" height="341" />
        </clipPath>
        <g data-lapisan="ab" className={nyala['ab'] ? undefined : 'ig-mati'}>
          <path d="M250 126 C285.2 126 304.4 148.7 314 185.02 C323.6 262.2 364 316.68 364 398.4 C364 498.28 312.7 580 250 580 C187.3 580 136 498.28 136 398.4 C136 316.68 176.4 262.2 186 185.02 C195.6 148.7 214.8 126 250 126 Z" fill="url(#ig2)" clipPath="url(#ig2-k1)" className="ig-ab-tepi" />
          <line x1="340" y1="261" x2="438" y2="300" className="ig-penghubung" />
          <g className="ig-pil ig-pil--ab">
            <rect x="449" y="277" width="66" height="46" rx="23" />
            <text x="482" y="309" textAnchor="middle">AB</text>
          </g>
        </g>
        <path d="M208 60 Q250 34 292 60 L287.8 132 L267.64 305 Q250 321 232.36 305 L212.2 132 Z" className="ig-gigi" />
        <g data-lapisan="mc" className={nyala['mc'] ? undefined : 'ig-mati'}>
          <circle cx="254" cy="450" r="26" className="ig-kanal" />
          <line x1="235" y1="450" x2="273" y2="450" className="ig-kanal-garis" />
          <line x1="280" y1="450" x2="438" y2="400" className="ig-penghubung" />
          <g className="ig-pil ig-pil--mc">
            <rect x="449" y="377" width="66" height="46" rx="23" />
            <text x="482" y="409" textAnchor="middle">MC</text>
          </g>
        </g>
      </g>
      <g data-slice="3" className={pilihan['slice'] === '3' ? undefined : 'ig-mati'}>
        <path d="M250 172 C288.5 172 309.5 192.4 320 225.04 C330.5 294.4 370 343.36 370 416.79999999999995 C370 506.56 316 580 250 580 C184 580 130 506.56 130 416.79999999999995 C130 343.36 169.5 294.4 180 225.04 C190.5 192.4 211.5 172 250 172 Z" className="ig-tulang" />
        <clipPath id="ig2-k2">
          <rect x="0" y="0" width="800" height="387" />
        </clipPath>
        <g data-lapisan="ab" className={nyala['ab'] ? undefined : 'ig-mati'}>
          <path d="M250 172 C288.5 172 309.5 192.4 320 225.04 C330.5 294.4 370 343.36 370 416.79999999999995 C370 506.56 316 580 250 580 C184 580 130 506.56 130 416.79999999999995 C130 343.36 169.5 294.4 180 225.04 C190.5 192.4 211.5 172 250 172 Z" fill="url(#ig2)" clipPath="url(#ig2-k2)" className="ig-ab-tepi" />
          <line x1="346" y1="307" x2="438" y2="300" className="ig-penghubung" />
          <g className="ig-pil ig-pil--ab">
            <rect x="449" y="277" width="66" height="46" rx="23" />
            <text x="482" y="309" textAnchor="middle">AB</text>
          </g>
        </g>
        <g data-lapisan="mc" className={nyala['mc'] ? undefined : 'ig-mati'}>
          <circle cx="250" cy="454" r="27" className="ig-kanal" />
          <line x1="230" y1="454" x2="270" y2="454" className="ig-kanal-garis" />
          <line x1="277" y1="454" x2="438" y2="400" className="ig-penghubung" />
          <g className="ig-pil ig-pil--mc">
            <rect x="449" y="377" width="66" height="46" rx="23" />
            <text x="482" y="409" textAnchor="middle">MC</text>
          </g>
        </g>
      </g>
      <g data-slice="4" className={pilihan['slice'] === '4' ? undefined : 'ig-mati'}>
        <path d="M250 126 C296.2 126 321.4 148.7 334 185.02 C346.6 262.2 376 316.68 376 398.4 C376 498.28 319.3 580 250 580 C180.7 580 124 498.28 124 398.4 C124 316.68 153.4 262.2 166 185.02 C178.60000000000002 148.7 203.8 126 250 126 Z" className="ig-tulang" />
        <clipPath id="ig2-k3">
          <rect x="0" y="0" width="800" height="341" />
        </clipPath>
        <g data-lapisan="ab" className={nyala['ab'] ? undefined : 'ig-mati'}>
          <path d="M250 126 C296.2 126 321.4 148.7 334 185.02 C346.6 262.2 376 316.68 376 398.4 C376 498.28 319.3 580 250 580 C180.7 580 124 498.28 124 398.4 C124 316.68 153.4 262.2 166 185.02 C178.60000000000002 148.7 203.8 126 250 126 Z" fill="url(#ig2)" clipPath="url(#ig2-k3)" className="ig-ab-tepi" />
          <line x1="360" y1="261" x2="438" y2="300" className="ig-penghubung" />
          <g className="ig-pil ig-pil--ab">
            <rect x="449" y="277" width="66" height="46" rx="23" />
            <text x="482" y="309" textAnchor="middle">AB</text>
          </g>
        </g>
        <path d="M188 60 Q250 34 312 60 L305.8 132 L294.64 300 Q276.04 310 262.4 170 Q250 156 237.6 170 Q223.96 310 205.36 300 L194.2 132 Z" className="ig-gigi" />
        <g data-lapisan="mc" className={nyala['mc'] ? undefined : 'ig-mati'}>
          <circle cx="244" cy="460" r="28" className="ig-kanal" />
          <line x1="223" y1="460" x2="265" y2="460" className="ig-kanal-garis" />
          <line x1="272" y1="460" x2="438" y2="400" className="ig-penghubung" />
          <g className="ig-pil ig-pil--mc">
            <rect x="449" y="377" width="66" height="46" rx="23" />
            <text x="482" y="409" textAnchor="middle">MC</text>
          </g>
        </g>
      </g>
      <g data-slice="5" className={pilihan['slice'] === '5' ? undefined : 'ig-mati'}>
        <path d="M250 120 C300.6 120 328.2 143 342 179.8 C355.8 258 382 313.2 382 396 C382 497.2 322.6 580 250 580 C177.39999999999998 580 118 497.2 118 396 C118 313.2 144.2 258 158 179.8 C171.8 143 199.4 120 250 120 Z" className="ig-tulang" />
        <clipPath id="ig2-k4">
          <rect x="0" y="0" width="800" height="335" />
        </clipPath>
        <g data-lapisan="ab" className={nyala['ab'] ? undefined : 'ig-mati'}>
          <path d="M250 120 C300.6 120 328.2 143 342 179.8 C355.8 258 382 313.2 382 396 C382 497.2 322.6 580 250 580 C177.39999999999998 580 118 497.2 118 396 C118 313.2 144.2 258 158 179.8 C171.8 143 199.4 120 250 120 Z" fill="url(#ig2)" clipPath="url(#ig2-k4)" className="ig-ab-tepi" />
          <line x1="368" y1="255" x2="438" y2="300" className="ig-penghubung" />
          <g className="ig-pil ig-pil--ab">
            <rect x="449" y="277" width="66" height="46" rx="23" />
            <text x="482" y="309" textAnchor="middle">AB</text>
          </g>
        </g>
        <path d="M184 54 Q250 28 316 54 L309.4 126 L297.52 296 Q277.72 306 263.2 164 Q250 150 236.8 164 Q222.28 306 202.48000000000002 296 L190.6 126 Z" className="ig-gigi" />
        <g data-lapisan="mc" className={nyala['mc'] ? undefined : 'ig-mati'}>
          <circle cx="238" cy="466" r="29" className="ig-kanal" />
          <line x1="216" y1="466" x2="260" y2="466" className="ig-kanal-garis" />
          <line x1="267" y1="466" x2="438" y2="400" className="ig-penghubung" />
          <g className="ig-pil ig-pil--mc">
            <rect x="449" y="377" width="66" height="46" rx="23" />
            <text x="482" y="409" textAnchor="middle">MC</text>
          </g>
        </g>
      </g>
      <rect x="540" y="22" width="248" height="212" rx="20" className="ig-kotak" />
      <path d="M741.9 59.1 A74 118 0 0 1 594.1 59.1" className="ig-rahang-tepi" strokeWidth="44" />
      <path d="M741.9 59.1 A74 118 0 0 1 594.1 59.1" className="ig-rahang-isi" strokeWidth="40" />
      <line x1="686" y1="136.3" x2="722.2" y2="179" className="ig-slice" />
      <g data-slice="1" className={pilihan['slice'] === '1' ? undefined : 'ig-mati'}>
        <line x1="686" y1="136.3" x2="722.2" y2="179" className="ig-slice ig-slice--aktif" />
        <circle cx="733.9" cy="192.7" r="16" className="ig-tanda ig-tanda--aktif" />
        <text x="733.9" y="199.7" textAnchor="middle" className="ig-nomor ig-nomor--kecil">1</text>
      </g>
      <line x1="694.4" y1="127.3" x2="739" y2="161.1" className="ig-slice" />
      <g data-slice="2" className={pilihan['slice'] === '2' ? undefined : 'ig-mati'}>
        <line x1="694.4" y1="127.3" x2="739" y2="161.1" className="ig-slice ig-slice--aktif" />
        <circle cx="753.4" cy="172" r="16" className="ig-tanda ig-tanda--aktif" />
        <text x="753.4" y="179" textAnchor="middle" className="ig-nomor ig-nomor--kecil">2</text>
      </g>
      <line x1="702.9" y1="113.4" x2="753.1" y2="138.1" className="ig-slice" />
      <g data-slice="3" className={pilihan['slice'] === '3' ? undefined : 'ig-mati'}>
        <line x1="702.9" y1="113.4" x2="753.1" y2="138.1" className="ig-slice ig-slice--aktif" />
        <circle cx="769.3" cy="146.1" r="16" className="ig-tanda ig-tanda--aktif" />
        <text x="769.3" y="153.1" textAnchor="middle" className="ig-nomor ig-nomor--kecil">3</text>
      </g>
      <line x1="710" y1="95.1" x2="763.6" y2="111" className="ig-slice" />
      <g data-slice="4" className={pilihan['slice'] === '4' ? undefined : 'ig-mati'}>
        <line x1="710" y1="95.1" x2="763.6" y2="111" className="ig-slice ig-slice--aktif" />
        <circle cx="780.9" cy="116.2" r="16" className="ig-tanda ig-tanda--aktif" />
        <text x="780.9" y="123.2" textAnchor="middle" className="ig-nomor ig-nomor--kecil">4</text>
      </g>
      <line x1="714.3" y1="75.2" x2="769.7" y2="83.4" className="ig-slice" />
      <g data-slice="5" className={pilihan['slice'] === '5' ? undefined : 'ig-mati'}>
        <line x1="714.3" y1="75.2" x2="769.7" y2="83.4" className="ig-slice ig-slice--aktif" />
        <circle cx="787.5" cy="86" r="16" className="ig-tanda ig-tanda--aktif" />
        <text x="787.5" y="93" textAnchor="middle" className="ig-nomor ig-nomor--kecil">5</text>
      </g>
    </svg>
  );
}
