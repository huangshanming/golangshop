/** 分类图标 3D 渐变配色 */
export const CAT_GRADIENTS = [
  'linear-gradient(145deg, #ffc285 0%, #ff8c42 45%, #e86a1a 100%)',
  'linear-gradient(145deg, #6ee8e2 0%, #3dbeb8 45%, #2a9d96 100%)',
  'linear-gradient(145deg, #ffe799 0%, #ffd166 45%, #e6b84d 100%)',
  'linear-gradient(145deg, #9ee4b8 0%, #73c991 45%, #52a872 100%)',
  'linear-gradient(145deg, #a8b8f5 0%, #7b8cde 45%, #5a6bc7 100%)',
  'linear-gradient(145deg, #f5aac8 0%, #e07a9f 45%, #c45a82 100%)',
  'linear-gradient(145deg, #9fd4ff 0%, #6bb8ff 45%, #4a9ae6 100%)',
  'linear-gradient(145deg, #d4b8ff 0%, #b088ea 45%, #9068c9 100%)',
]

export function catGradient(index: number) {
  return CAT_GRADIENTS[index % CAT_GRADIENTS.length]
}
