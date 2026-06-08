/** H5 固定底栏占位（Teleport 到 body 后页面需预留） */
export const H5_TAB_BAR_HEIGHT = 'calc(50px + env(safe-area-inset-bottom, 0px))'

/** 切换 Tab 后同步 custom-tab-bar 选中态（小程序端） */
export function syncTabBarSelected() {
  const pages = getCurrentPages()
  const page = pages[pages.length - 1] as { getTabBar?: () => { updateSelected?: () => void } }
  page?.getTabBar?.()?.updateSelected?.()
}
