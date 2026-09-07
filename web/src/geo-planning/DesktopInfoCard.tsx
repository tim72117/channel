import { X } from 'lucide-react'
import type { ReactNode } from 'react'
import styles from './DesktopInfoCard.module.css'

// DesktopInfoCard:桌面版浮動資訊卡共用外框(定位/避讓/關閉鍵/捲動
// 容器),PlacePanel/AttractionInfoPanel 的內容(照片/名稱/簡介/按鈕組)
// 透過 children 放進來——見 DesktopInfoCard.module.css 開頭對「為什麼
// 從各自獨立一份改成共用」的完整說明。這個元件不知道 content 的形狀,
// 純粹負責外框視覺與 shiftBy/style 這兩種定位輸入的套用邏輯。
export function DesktopInfoCard({
  onClose,
  shiftBy,
  style,
  children,
}: {
  onClose: () => void
  // shiftBy/style:理由與既有用法完全同 PlacePanel.tsx 對應 prop 的
  // 說明——shiftBy 是呼叫端依飯店側欄/對話小匡是否佔用右緣算出的三段式
  // 固定值,style 是「附近景點/level4-5 地標」並存卡片需要的動態 right
  // 位移逃生艙,兩者理論上不會同時傳(呼叫端目前的用法各自只用其中一種),
  // 但同時傳入時 style 的 CSS 優先序高於 class,以 style 為準。
  shiftBy?: 'none' | 'hotel' | 'chat'
  style?: React.CSSProperties
  children: ReactNode
}) {
  const shiftClass = shiftBy === 'chat' ? ` ${styles.shiftedChat}` : shiftBy === 'hotel' ? ` ${styles.shiftedHotel}` : ''
  return (
    <div className={`${styles.panel}${shiftClass}`} style={style}>
      <button type="button" className={styles.closeBtn} onClick={onClose} title="關閉">
        <X size={16} strokeWidth={2} />
      </button>
      <div className={styles.body}>{children}</div>
    </div>
  )
}
