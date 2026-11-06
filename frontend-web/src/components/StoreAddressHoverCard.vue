<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  HoverCardRoot,
  HoverCardTrigger,
  HoverCardPortal,
  HoverCardContent,
  HoverCardArrow,
} from 'reka-ui'
import {
  MapPin,
  ExternalLink,
  Copy,
  Check,
  Phone,
  Navigation,
  Store,
  Building2,
} from 'lucide-vue-next'
import { toast } from 'vue-sonner'

const props = withDefaults(
  defineProps<{
    address?: string | null
    shopName?: string | null
    province?: string | null
    regency?: string | null
    district?: string | null
    village?: string | null
    phone?: string | null
    latitude?: number | string | null
    longitude?: number | string | null
    salesName?: string | null
    salesBranch?: string | null
    side?: 'top' | 'right' | 'bottom' | 'left'
    align?: 'start' | 'center' | 'end'
    openDelay?: number
    closeDelay?: number
    clampLines?: number
    customClass?: string
    iconClass?: string
  }>(),
  {
    address: '',
    shopName: '',
    side: 'top',
    align: 'start',
    openDelay: 120,
    closeDelay: 150,
    clampLines: 1,
    customClass: '',
    iconClass: 'w-3 h-3',
  },
)

const isCopied = ref(false)

const formattedFullAddress = computed(() => {
  const parts: string[] = []
  if (props.address && props.address.trim()) {
    parts.push(props.address.trim())
  }
  const adminParts = [props.village, props.district, props.regency, props.province].filter(
    (p) => !!p && p.trim() !== '',
  )
  if (adminParts.length > 0) {
    const adminStr = adminParts.join(', ')
    // Avoid duplicate if address already contains the administrative parts
    if (!props.address || !props.address.toLowerCase().includes(adminParts[0]?.toLowerCase() || '')) {
      parts.push(adminStr)
    }
  }
  return parts.join(', ') || props.address || 'Alamat tidak tersedia'
})

const displayAddress = computed(() => {
  if (props.address && props.address.trim()) {
    return props.address.trim()
  }
  const adminParts = [props.village, props.district, props.regency, props.province].filter(
    (p) => !!p && p.trim() !== '',
  )
  if (adminParts.length > 0) {
    return adminParts.join(', ')
  }
  if (hasGpsCoordinates.value) {
    return `Koordinat (${props.latitude}, ${props.longitude})`
  }
  return 'Alamat tidak tersedia'
})

const hasGpsCoordinates = computed(() => {
  const lat = Number(props.latitude)
  const lng = Number(props.longitude)
  return !isNaN(lat) && !isNaN(lng) && (lat !== 0 || lng !== 0)
})

const googleMapsUrl = computed(() => {
  if (hasGpsCoordinates.value) {
    return `https://www.google.com/maps/search/?api=1&query=${props.latitude},${props.longitude}`
  }
  if (props.address && props.address.trim()) {
    const query = encodeURIComponent(
      [props.shopName, props.address, props.regency, props.province].filter(Boolean).join(', '),
    )
    return `https://www.google.com/maps/search/?api=1&query=${query}`
  }
  return null
})

const copyAddress = async (e?: Event) => {
  if (e) e.stopPropagation()
  const textToCopy = formattedFullAddress.value
  if (!textToCopy || textToCopy === 'Alamat tidak tersedia') return
  try {
    await navigator.clipboard.writeText(textToCopy)
    isCopied.value = true
    toast.success('Alamat toko berhasil disalin ke clipboard!')
    setTimeout(() => {
      isCopied.value = false
    }, 2000)
  } catch {
    toast.error('Gagal menyalin alamat')
  }
}
</script>

<template>
  <HoverCardRoot :open-delay="openDelay" :close-delay="closeDelay">
    <HoverCardTrigger as-child>
      <slot>
        <div
          class="inline-flex items-center gap-1 text-slate-500 hover:text-primary transition-colors cursor-pointer group/addr text-left"
          :class="customClass"
        >
          <MapPin
            :class="[iconClass, 'shrink-0 text-slate-400 group-hover/addr:text-primary transition-colors']"
          />
          <span
            class="font-medium hover:underline decoration-dotted underline-offset-2 transition-all"
            :class="clampLines === 1 ? 'line-clamp-1' : clampLines === 2 ? 'line-clamp-2' : ''"
          >
            {{ displayAddress }}
          </span>
        </div>
      </slot>
    </HoverCardTrigger>

    <HoverCardPortal>
      <HoverCardContent
        :side="side"
        :align="align"
        :side-offset="8"
        class="z-[99999] w-84 max-w-[92vw] rounded-2xl bg-white p-4 shadow-2xl border border-slate-200/90 outline-none animate-in fade-in zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 duration-150 text-slate-800"
      >
        <div class="space-y-3">
          <!-- Header -->
          <div class="flex items-start justify-between gap-2 border-b border-slate-100 pb-2.5">
            <div class="flex items-center gap-2 min-w-0">
              <div class="w-7 h-7 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                <Store class="w-4 h-4" />
              </div>
              <div class="min-w-0">
                <h4 class="text-xs font-black text-slate-900 leading-tight truncate">
                  {{ shopName || 'Detail Toko' }}
                </h4>
                <p v-if="salesBranch" class="text-[10px] text-slate-400 font-medium truncate">
                  Cabang: {{ salesBranch }}
                </p>
              </div>
            </div>

            <span
              class="text-[9px] font-black uppercase tracking-wider bg-slate-100 text-slate-600 px-2 py-0.5 rounded-full border border-slate-200/60 shrink-0"
            >
              Lokasi Toko
            </span>
          </div>

          <!-- Full Address Box -->
          <div class="bg-slate-50/90 rounded-xl p-2.5 border border-slate-100 space-y-2">
            <div class="flex items-start gap-2">
              <MapPin class="w-3.5 h-3.5 text-rose-500 shrink-0 mt-0.5" />
              <div class="text-xs text-slate-700 leading-relaxed font-medium break-words">
                {{ formattedFullAddress }}
              </div>
            </div>

            <!-- Administrative Badges (if available) -->
            <div
              v-if="village || district || regency || province"
              class="flex flex-wrap gap-1 pt-1.5 border-t border-slate-200/60 text-[10px]"
            >
              <span
                v-if="village"
                class="bg-white px-2 py-0.5 rounded-md border border-slate-200 text-slate-600 font-semibold"
                title="Kelurahan / Desa"
              >
                Ds. {{ village }}
              </span>
              <span
                v-if="district"
                class="bg-white px-2 py-0.5 rounded-md border border-slate-200 text-slate-600 font-semibold"
                title="Kecamatan"
              >
                Kec. {{ district }}
              </span>
              <span
                v-if="regency"
                class="bg-white px-2 py-0.5 rounded-md border border-slate-200 text-slate-600 font-semibold"
                title="Kabupaten / Kota"
              >
                {{ regency }}
              </span>
              <span
                v-if="province"
                class="bg-white px-2 py-0.5 rounded-md border border-slate-200 text-slate-600 font-semibold"
                title="Provinsi"
              >
                Prov. {{ province }}
              </span>
            </div>
          </div>

          <!-- Phone & Sales Info -->
          <div
            v-if="phone || salesName"
            class="grid grid-cols-1 gap-1.5 text-[11px] text-slate-600 px-0.5"
          >
            <div v-if="phone" class="flex items-center gap-1.5">
              <Phone class="w-3 h-3 text-emerald-600 shrink-0" />
              <span>Telp: <strong class="text-slate-800">{{ phone }}</strong></span>
            </div>
            <div v-if="salesName" class="flex items-center gap-1.5">
              <Building2 class="w-3 h-3 text-blue-500 shrink-0" />
              <span>Sales PIC: <strong class="text-slate-800">{{ salesName }}</strong></span>
            </div>
          </div>

          <!-- Action Buttons -->
          <div class="flex items-center gap-2 pt-1 border-t border-slate-100">
            <!-- Copy Address Button -->
            <button
              type="button"
              @click.stop="copyAddress"
              class="flex-1 inline-flex items-center justify-center gap-1.5 text-xs font-bold px-3 py-1.5 rounded-lg border transition-all cursor-pointer shadow-xs"
              :class="
                isCopied
                  ? 'bg-emerald-50 text-emerald-700 border-emerald-300'
                  : 'bg-white hover:bg-slate-50 text-slate-700 border-slate-200 hover:border-slate-300'
              "
            >
              <Check v-if="isCopied" class="w-3.5 h-3.5 text-emerald-600" />
              <Copy v-else class="w-3.5 h-3.5 text-slate-500" />
              <span>{{ isCopied ? 'Tersalin' : 'Salin Alamat' }}</span>
            </button>

            <!-- Google Maps Button -->
            <a
              v-if="googleMapsUrl"
              :href="googleMapsUrl"
              target="_blank"
              rel="noopener noreferrer"
              @click.stop
              class="flex-1 inline-flex items-center justify-center gap-1.5 text-xs font-bold px-3 py-1.5 rounded-lg bg-primary hover:bg-primary/90 text-white transition-all shadow-xs"
            >
              <Navigation class="w-3.5 h-3.5" />
              <span>Google Maps</span>
              <ExternalLink class="w-2.5 h-2.5 opacity-80" />
            </a>
          </div>
        </div>

        <HoverCardArrow class="fill-white drop-shadow-sm" />
      </HoverCardContent>
    </HoverCardPortal>
  </HoverCardRoot>
</template>
