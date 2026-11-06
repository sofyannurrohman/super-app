<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { watchDebounced } from '@vueuse/core'
import {
  FlexRender,
  getCoreRowModel,
  useVueTable,
  getPaginationRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  type SortingState,
} from '@tanstack/vue-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const props = defineProps<{
  data: any[]
  columns: any[]
  isLoading?: boolean
  serverSide?: boolean
  totalItems?: number
  currentPage?: number
  pageSize?: number
}>()

const emit = defineEmits(['update:page', 'update:limit', 'search'])

// --- Column Normalization ---
// Converts simple {key, label} columns to TanStack format internally
const normalizedColumns = computed(() => {
  if (!props.columns) return []

  return props.columns.map((col) => {
    // If it's already in TanStack format, keep it
    if (col.accessorKey || col.id || col.header) return col

    // Otherwise, convert from legacy format
    const newCol: any = {
      accessorKey: col.key,
      id: col.key,
      header: col.label,
    }

    // Handle custom render if provided in legacy format
    if (col.render) {
      newCol.cell = (info: any) =>
        h('div', { innerHTML: col.render(info.getValue(), info.row.original) })
    } else if (col.type === 'datetime') {
      newCol.cell = (info: any) => {
        const val = info.getValue()
        return val ? new Date(val).toLocaleString('id-ID') : '-'
      }
    } else if (col.type === 'currency') {
      newCol.cell = (info: any) => {
        const val = info.getValue()
        return val ? `Rp ${Number(val).toLocaleString('id-ID')}` : '-'
      }
    }

    return newCol
  })
})

const sorting = ref<SortingState>([])
const filterText = ref('')
const globalFilter = ref('')

// Performance: Debounce search to prevent laggy typing and redundant re-renders
watchDebounced(
  filterText,
  (val) => {
    if (props.serverSide) {
      emit('search', val)
    } else {
      globalFilter.value = val
    }
  },
  { debounce: 300 },
)

const table = useVueTable({
  get data() {
    return props.data || []
  },
  get columns() {
    return normalizedColumns.value
  },
  state: {
    get sorting() {
      return sorting.value
    },
    get globalFilter() {
      return globalFilter.value
    },
    get pagination() {
      if (props.serverSide && props.currentPage && props.pageSize) {
        return {
          pageIndex: props.currentPage - 1,
          pageSize: Number(props.pageSize),
        }
      }
      return undefined
    },
  },
  manualPagination: props.serverSide,
  get pageCount() {
    if (props.serverSide) {
      const size = Number(props.pageSize) || 10
      return Math.ceil((props.totalItems || 0) / size)
    }
    return undefined
  },
  manualFiltering: props.serverSide,
  onSortingChange: (updaterOrValue) => {
    sorting.value =
      typeof updaterOrValue === 'function' ? updaterOrValue(sorting.value) : updaterOrValue
  },
  onGlobalFilterChange: (setFilter) => {
    globalFilter.value = setFilter
  },
  getCoreRowModel: getCoreRowModel(),
  getPaginationRowModel: getPaginationRowModel(),
  getSortedRowModel: getSortedRowModel(),
  getFilteredRowModel: getFilteredRowModel(),
  initialState: {
    pagination: {
      pageSize: 10,
    },
  },
})
// --- Smart Pagination ---
// Builds a sliding window of page numbers with ellipsis for large page counts.
// Always shows first page, last page, and up to MAX_VISIBLE_PAGES pages around the active page.
const MAX_VISIBLE_PAGES = 5

const currentPageNumber = computed(() => {
  if (props.serverSide) return props.currentPage || 1
  return table.getState().pagination.pageIndex + 1
})

const visiblePages = computed<(number | '...')[]>(() => {
  const total = table.getPageCount()
  if (!total || total <= 0) return []

  // Show all pages if total fits within window
  if (total <= MAX_VISIBLE_PAGES + 2) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }

  const current = currentPageNumber.value
  const half = Math.floor((MAX_VISIBLE_PAGES - 2) / 2)
  const winStart = Math.max(2, current - half)
  const winEnd = Math.min(total - 1, current + half)

  const pages: (number | '...')[] = [1]

  if (winStart > 2) pages.push('...')
  for (let i = winStart; i <= winEnd; i++) pages.push(i)
  if (winEnd < total - 1) pages.push('...')

  pages.push(total)
  return pages
})

const isActivePage = (page: number | '...') => {
  if (page === '...') return false
  return page === currentPageNumber.value
}

const goToPage = (page: number | '...') => {
  if (page === '...') return
  if (props.serverSide) {
    emit('update:page', page)
  } else {
    table.setPageIndex((page as number) - 1)
  }
}
</script>

<template>
  <div
    class="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden flex flex-col w-full"
  >
    <!-- Search / Top Actions -->

    <div
      class="p-5 border-b border-gray-100 flex flex-col sm:flex-row items-center justify-between gap-4"
    >
      <div class="flex items-center gap-3">
        <slot name="headerTitle" />
      </div>
      <div class="flex items-center gap-2">
        <div class="relative w-64">
          <svg
            class="absolute left-3 top-2.5 w-4 h-4 text-gray-400 z-10"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
          <Input
            v-model="filterText"
            type="text"
            placeholder="Cari data..."
            class="pl-9 bg-gray-50 focus-visible:ring-primary h-10 rounded-xl"
          />
        </div>
        <slot name="headerActions" />
      </div>
    </div>

    <!-- Table -->
    <div class="overflow-x-auto min-h-[400px]">
      <Table>
        <TableHeader
          class="bg-[#fafbfc] uppercase text-[11px] font-bold text-gray-500 tracking-wider"
        >
          <TableRow
            v-for="headerGroup in table.getHeaderGroups()"
            :key="headerGroup.id"
            class="hover:bg-transparent border-b border-gray-100"
          >
            <TableHead
              v-for="header in headerGroup.headers"
              :key="header.id"
              class="cursor-pointer select-none group h-12"
              @click="header.column.getToggleSortingHandler()?.($event)"
            >
              <div class="flex items-center gap-1">
                <slot :name="`header-${header.column.id}`" :header="header">
                  <FlexRender
                    v-if="!header.isPlaceholder"
                    :render="header.column.columnDef.header"
                    :props="header.getContext()"
                  />
                </slot>
                <!-- Sorting Indicator -->
                <span v-if="header.column.getCanSort()">
                  <svg
                    v-if="header.column.getIsSorted() === 'asc'"
                    class="w-3 h-3 text-primary"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="3"
                      d="M5 15l7-7 7 7"
                    ></path>
                  </svg>
                  <svg
                    v-else-if="header.column.getIsSorted() === 'desc'"
                    class="w-3 h-3 text-primary"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="3"
                      d="M19 9l-7 7-7-7"
                    ></path>
                  </svg>
                  <svg
                    v-else
                    class="w-3 h-3 text-gray-300 group-hover:text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M8 9l4-4 4 4m0 6l-4 4-4-4"
                    ></path>
                  </svg>
                </span>
              </div>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody class="relative text-[14px]">
          <!-- Performance: Skeleton Loading Blocks -->
          <template v-if="isLoading">
            <TableRow v-for="i in 5" :key="'skeleton-' + i" class="border-b border-gray-50">
              <TableCell v-for="col in columns" :key="'col-' + i">
                <div class="h-4 bg-gray-100 rounded-full animate-pulse w-full"></div>
              </TableCell>
            </TableRow>
          </template>

          <template v-else-if="table.getRowModel().rows.length === 0">
            <TableRow>
              <TableCell :colspan="columns.length" class="py-20 text-center">
                <div class="flex flex-col items-center gap-2">
                  <div
                    class="w-12 h-12 rounded-full bg-gray-50 flex items-center justify-center text-gray-300"
                  >
                    <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
                      />
                    </svg>
                  </div>
                  <p class="text-gray-500 font-medium">Tidak ada data ditemukan</p>
                </div>
              </TableCell>
            </TableRow>
          </template>

          <template v-else>
            <TableRow
              v-for="row in table.getRowModel().rows"
              :key="row.id"
              class="hover:bg-gray-50/50 transition-all border-b border-gray-50"
            >
              <TableCell v-for="cell in row.getVisibleCells()" :key="cell.id" class="py-4">
                <slot :name="`cell-${cell.column.id}`" :row="row.original" :cell="cell">
                  <FlexRender :render="cell.column.columnDef.cell" :props="cell.getContext()" />
                </slot>
              </TableCell>
            </TableRow>
          </template>
        </TableBody>
      </Table>
    </div>

    <!-- Pagination -->
    <div
      class="px-4 py-3 border-t border-gray-100 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
    >
      <!-- Info teks -->
      <span class="text-xs sm:text-[13px] text-gray-400 shrink-0">
        <template v-if="props.serverSide">
          Halaman {{ props.currentPage || 1 }} dari {{ table.getPageCount() || 1 }} &bull;
          {{ props.totalItems || 0 }} data
        </template>
        <template v-else>
          Halaman {{ currentPageNumber }} dari {{ table.getPageCount() || 1 }} &bull;
          {{ table.getFilteredRowModel().rows.length }} data
        </template>
      </span>

      <!-- Kontrol navigasi — scrollable sebagai fallback di mobile -->
      <div class="overflow-x-auto pb-0.5 -mx-1 px-1">
        <div class="flex items-center gap-1 w-max">
          <!-- Prev -->
          <Button
            variant="outline"
            size="icon"
            class="h-8 w-8 shrink-0 rounded-lg border-gray-200 text-gray-600 hover:bg-gray-50 hover:border-gray-300 disabled:opacity-40"
            :disabled="
              props.serverSide ? (props.currentPage || 1) <= 1 : !table.getCanPreviousPage()
            "
            @click="
              props.serverSide
                ? emit('update:page', (props.currentPage || 1) - 1)
                : table.previousPage()
            "
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
            </svg>
          </Button>

          <!-- Page buttons (smart: ellipsis window) -->
          <template v-for="(page, idx) in visiblePages" :key="idx">
            <!-- Ellipsis -->
            <span
              v-if="page === '...'"
              class="flex items-center justify-center h-8 w-7 text-gray-400 text-xs select-none"
            >
              &hellip;
            </span>

            <!-- Page number button -->
            <button
              v-else
              class="flex items-center justify-center h-8 min-w-[2rem] px-1.5 rounded-lg text-[13px] font-medium transition-all duration-150 border shrink-0"
              :class="
                isActivePage(page)
                  ? 'bg-primary text-white border-primary shadow-sm shadow-primary/25 scale-105'
                  : 'bg-white text-gray-600 border-gray-200 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-800'
              "
              @click="goToPage(page)"
            >
              {{ page }}
            </button>
          </template>

          <!-- Next -->
          <Button
            variant="outline"
            size="icon"
            class="h-8 w-8 shrink-0 rounded-lg border-gray-200 text-gray-600 hover:bg-gray-50 hover:border-gray-300 disabled:opacity-40"
            :disabled="
              props.serverSide
                ? (props.currentPage || 1) >= table.getPageCount()
                : !table.getCanNextPage()
            "
            @click="
              props.serverSide
                ? emit('update:page', (props.currentPage || 1) + 1)
                : table.nextPage()
            "
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

