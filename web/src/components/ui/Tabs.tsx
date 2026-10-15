import React from 'react'
import { UnderlineNav, type NavTabItem } from './UnderlineNav'

export type TabItem = NavTabItem

export interface TabsProps {
  tabs: TabItem[]
  activeTab: string
  onChange: (id: string) => void
  className?: string
}

export const Tabs: React.FC<TabsProps> = ({ tabs, activeTab, onChange, className }) => {
  return <UnderlineNav tabs={tabs} activeTab={activeTab} onChange={onChange} className={className} />
}
