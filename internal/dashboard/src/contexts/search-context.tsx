import { createContext, useContext, useState, type ReactNode } from "react"

type SearchContextValue = {
  query: string
  setQuery: (query: string) => void
}

const SearchContext = createContext<SearchContextValue | null>(null)

export function SearchProvider({ children }: Readonly<{ children: ReactNode }>) {
  const [query, setQuery] = useState("")
  return <SearchContext value={{ query, setQuery }}>{children}</SearchContext>
}

export function useDashboardSearch() {
  const context = useContext(SearchContext)
  if (!context) throw new Error("useDashboardSearch must be used within SearchProvider")
  return context
}
