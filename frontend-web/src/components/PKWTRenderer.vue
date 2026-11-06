<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { brand } from '@/config/brand'

const props = defineProps<{
  data: {
    documentNumber: string
    agreementDay: string
    agreementDate: string | undefined
    agreementPlace: string
    representativeName: string
    representativePosition: string
    employee: any
    startDate: string | undefined
    endDate: string | undefined
    jobPosition: string
    basicSalary: number
    allowances: number
    gmMarketingName?: string
    employeeSignatureUrl?: string
    hr_signature_url?: string
    probation_months?: number
  }
}>()

const authStore = useAuthStore()
const companyName = computed(() => props.data.employee?.company?.name || authStore.user?.company_name || 'Perusahaan')
const companyAddress = computed(() => props.data.employee?.company?.address || authStore.user?.company_address || '')
const companyLogoUrl = computed(() => props.data.employee?.company?.logo_url || authStore.user?.company_logo_url || brand.logoUrl)

const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return '....................'
  const date = new Date(dateStr)
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

const getYear = (dateStr: string | undefined) => {
  if (!dateStr) return '2026'
  return new Date(dateStr).getFullYear()
}

const formatCurrency = (amount: number) => {
  if (!amount) return '0'
  return amount.toLocaleString('id-ID')
}

const getSignatureFullUrl = (url: string | undefined) => {
  if (!url) return ''
  if (url.startsWith('http')) return url
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || ''
  try {
    const origin = new URL(apiBaseUrl).origin
    return `${origin}${url.startsWith('/') ? url : '/' + url}`
  } catch (e) {
    return url
  }
}
</script>

<template>
  <div
    class="pkwt-container bg-white text-[11pt] leading-[1.4] font-serif text-black min-h-screen mx-auto max-w-[21cm] shadow-lg print:shadow-none print:p-0 p-4"
  >
    <!-- PAGE 1: COVER PAGE -->
    <div class="page-container p-[1cm] flex flex-col items-center">
      <h1 class="font-bold text-[18pt] mb-12 uppercase text-center">
        PERJANJIAN KERJA WAKTU TERTENTU
      </h1>

      <div class="mb-12 flex justify-center">
        <img :src="companyLogoUrl" alt="Logo perusahaan" class="h-32 object-contain" />
      </div>

      <div class="w-[400px] mb-12">
        <table class="w-full border-collapse border border-black text-[10pt]">
          <tr>
            <td class="border border-black px-2 py-1 w-[80px]">Nomor</td>
            <td class="border border-black px-2 py-1 w-[20px] text-center">:</td>
            <td class="border border-black px-2 py-1">
              {{ props.data.documentNumber || '....' }}/DEMO-PKWT/{{
                getYear(props.data.agreementDate)
              }}
            </td>
          </tr>
          <tr>
            <td class="border border-black px-2 py-1">Tanggal</td>
            <td class="border border-black px-2 py-1 text-center">:</td>
            <td class="border border-black px-2 py-1">
              {{ formatDate(props.data.agreementDate) }}
            </td>
          </tr>
          <tr>
            <td class="border border-black px-2 py-1">TMT</td>
            <td class="border border-black px-2 py-1 text-center">:</td>
            <td class="border border-black px-2 py-1">{{ formatDate(props.data.startDate) }}</td>
          </tr>
        </table>
      </div>

      <p class="text-center mb-2">Oleh dan diantara:</p>

      <div class="w-full">
        <table class="w-full border-collapse border border-black text-[10pt]">
          <tr class="font-bold uppercase text-center bg-gray-50">
            <td colspan="2" class="border border-black py-2 w-1/2">PERUSAHAAN</td>
            <td colspan="2" class="border border-black py-2 w-1/2">KARYAWAN</td>
          </tr>
          <tr>
            <td class="border border-black px-2 py-1 align-top w-[60px]">Nama</td>
            <td class="border border-black px-2 py-1 font-bold">: {{ companyName }}</td>
            <td class="border border-black px-2 py-1 align-top w-[60px]">Nama</td>
            <td class="border border-black px-2 py-1">
              : {{ props.data.employee?.first_name }} {{ props.data.employee?.last_name }}
            </td>
          </tr>
          <tr class="h-24">
            <td class="border border-black px-2 py-1 align-top">Alamat</td>
            <td class="border border-black px-2 py-1 text-[9pt] align-top">
              : {{ companyAddress || '....................' }}
            </td>
            <td class="border border-black px-2 py-1 align-top">Alamat</td>
            <td class="border border-black px-2 py-1 text-[9pt] align-top">
              : {{ props.data.employee?.address_ktp || '....................' }}
            </td>
          </tr>
        </table>
      </div>

      <div class="mt-auto flex justify-end">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 2: PARTIES DETAIL -->
    <div class="page-container p-[1cm]">
      <div class="text-center mb-4 space-y-1">
        <h1 class="font-bold underline uppercase text-[12pt]">PERJANJIAN KERJA WAKTU TERTENTU</h1>
        <p class="font-bold uppercase">ANTARA {{ companyName }}</p>
        <p class="font-bold uppercase">DENGAN Sdr/i:</p>
        <p class="mt-4">
          NOMOR : {{ props.data.documentNumber || '....' }}/DEMO-PKWT/{{
            getYear(props.data.agreementDate)
          }}
        </p>
      </div>

      <div class="mb-2">
        <p class="font-bold">PERJANJIAN KERJA WAKTU TERTENTU ini dibuat pada:</p>
        <div class="grid grid-cols-[120px_20px_1fr] mt-1">
          <span>Hari</span><span>:</span
          ><span>{{ props.data.agreementDay || '....................' }}</span> <span>Tanggal</span
          ><span>:</span><span>{{ formatDate(props.data.agreementDate) }}</span>
          <span>Bertempat di</span><span>:</span
          ><span>{{ props.data.agreementPlace || '....................' }}</span>
        </div>
      </div>

      <p class="mb-2 font-bold">("<span class="uppercase">PERJANJIAN</span>") Oleh dan diantara:</p>

      <!-- Pihak Pertama -->
      <div class="grid grid-cols-[40px_140px_20px_1fr] mb-4 gap-y-1">
        <span class="font-bold">(1)</span><span>Nama</span><span>:</span
        ><span class="font-bold uppercase">{{ companyName }}</span> <span></span
        ><span>Alamat</span><span>:</span>
        <div class="grid grid-cols-[160px_20px_1fr]">
          <span>Alamat Perusahaan</span><span>:</span><span>{{ companyAddress || '....................' }}</span>
        </div>
        <div class="col-span-4 mt-2 pl-[40px]">
          <p class="text-justify">
            Dalam hal ini diwakili oleh
            <span class="font-bold">{{ props.data.representativeName }}</span> selaku
            <span class="font-bold uppercase">{{ props.data.representativePosition }}</span> di
            {{ companyName }} dan menurut jabatannya bertindak untuk dan atas nama
            {{ companyName }} sebagai perusahaan pemberi kerja, oleh dan karenanya sah dan
            berwenang untuk bertindak untuk dan atas nama perusahaan (selanjutnya dalam perjanjian ini
            disebut "<span class="font-bold uppercase">PIHAK PERTAMA</span>").
          </p>
        </div>
      </div>

      <!-- Pihak Kedua -->
      <div class="grid grid-cols-[40px_140px_20px_1fr] mb-4 gap-y-1">
        <span class="font-bold">(2)</span><span>Nama</span><span>:</span
        ><span class="font-bold uppercase"
          >{{ props.data.employee?.first_name }} {{ props.data.employee?.last_name }}</span
        >
        <span></span><span>Tempat/Tanggal Lahir</span><span>:</span
        ><span
          >{{ props.data.employee?.birth_place }},
          {{ formatDate(props.data.employee?.birth_date) }}</span
        >
        <span></span><span>Jenis Kelamin</span><span>:</span
        ><span>{{ props.data.employee?.gender }}</span> <span></span><span>Alamat</span
        ><span>:</span>
        <div class="grid grid-cols-[160px_20px_1fr]">
          <span>Nama/Nomor Jalan</span><span>:</span
          ><span>{{ props.data.employee?.address_ktp || '....................' }}</span>
          <span>RT/RW</span><span>:</span><span>....................</span> <span>Kelurahan</span
          ><span>:</span><span>....................</span> <span>Kecamatan</span><span>:</span
          ><span>....................</span> <span>Kabupaten/Kota</span><span>:</span
          ><span>....................</span> <span>Provinsi</span><span>:</span
          ><span>....................</span>
        </div>
        <span></span><span>NIK</span><span>:</span
        ><span>{{ props.data.employee?.identity_number || '....................' }}</span>
        <div class="col-span-4 mt-2 pl-[40px]">
          <p class="text-justify">
            (selanjutnya dalam perjanjian ini disebut "<span class="font-bold uppercase"
              >PIHAK KEDUA</span
            >" yang merupakan Penerima Kerja/Karyawan).
          </p>
        </div>
      </div>

      <div class="mt-auto flex justify-end">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 3: PREAMBLE & PASAL 1 -->
    <div class="page-container p-[1cm]">
      <p class="mb-4 mt-4 text-justify">
        Perusahaan dan Karyawan secara bersama-sama selanjutnya disebut sebagai "<span
          class="font-bold"
          >PARA PIHAK</span
        >". <span class="font-bold">PARA PIHAK</span> dengan ini terlebih dahulu menerangkan hal-hal
        sebagai berikut:
      </p>

      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            >Bahwa berdasarkan kebutuhan akan tenaga kerja dan pelaksanaan proses perekrutan
            Karyawan, <span class="font-bold">PIHAK KEDUA</span> telah mengikuti proses tersebut
            secara sukarela dengan hasil memenuhi persyaratan umum and/atau khusus, maka
            <span class="font-bold">PIHAK PERTAMA</span> menerima
            <span class="font-bold">PIHAK KEDUA</span> sebagai karyawan dan
            <span class="font-bold">PIHAK KEDUA</span> menerima maksud tersebut;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            >Bahwa dengan ini <span class="font-bold">PIHAK PERTAMA</span> dan
            <span class="font-bold">PIHAK KEDUA</span> secara bersama-sama setuju untuk mengadakan
            Perjanjian Kerja yang dilandasi azas kesetaraan dan perlindungan kepentingan profesional
            dengan ketentuan dan syarat-syarat untuk ditaati dan dilaksanakan oleh
            <span class="font-bold">PARA PIHAK</span> seperti tercantum dalam perjanjian ini.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>3.</span>
          <span
            >Bahwa, Perusahaan dalam rangka menjalankan kegiatan usahanya bermaksud untuk
            mempekerjakan Karyawan secara tidak tetap berdasarkan
            <span class="font-bold">Perjanjian Kerja Waktu Tertentu (PKWT)</span> dan Karyawan
            sepakat untuk bekerja secara tidak tetap berdasarkan
            <span class="font-bold">Perjanjian Kerja Waktu Tertentu (PKWT)</span> bagi dan untuk
            kepentingan Perusahaan, yang ketentuan-ketentuan dan syarat-syaratnya sebagaimana diatur
            dan ditentukan berdasarkan Perjanjian ini.</span
          >
        </div>
      </div>

      <p class="mb-3 text-justify">
        Berdasarkan hal-hal tersebut diatas dan dengan iktikad baik,
        <span class="font-bold">PARA PIHAK</span> dengan ini sepakat untuk saling mengikatkan diri
        dalam Perjanjian ini dengan syarat-syarat dan ketentuan-ketentuan sebagai berikut :
      </p>

      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 1</p>
        <p class="font-bold uppercase">PENGERTIAN UMUM</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            ><span class="font-bold">PIHAK PERTAMA</span> adalah Perusahaan dalam hal ini sebagai
            pemberi kerja.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            ><span class="font-bold">PIHAK KEDUA</span> adalah Karyawan dalam hal ini sebagai
            penerima kerja.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>3.</span>
          <span
            ><span class="font-bold">Pekerja Waktu Tertentu</span> adalah orang yang dipekerjakan
            karena keahliannya dan dibutuhkan Perusahaan dalam jangka waktu tertentu diangkat
            sebagai pegawai tidak tetap berdasarkan Surat Perjanjian Kerja Waktu Tertentu, yang hak
            dan kewajibannya diatur dalam surat perjanjian kerja waktu tertentu.</span
          >
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 3: Pasal 2 & 3 -->
    <div class="page-container p-[1cm]">
      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 2</p>
        <p class="font-bold uppercase">JABATAN, TUGAS DAN KOMITMEN</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            ><span class="font-bold">PIHAK PERTAMA</span> akan memperkerjakan
            <span class="font-bold">PIHAK KEDUA</span> di {{ companyName }} beralamat di {{ companyAddress }}
            sesuai dengan kompetensi kerjanya, sebagai :
            <span class="font-bold underline">{{ props.data.jobPosition }}</span> terhitung mulai
            tanggal <span class="font-bold underline">{{ formatDate(props.data.startDate) }}</span
            >.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            ><span class="font-bold">Kompetensi Kerja</span> adalah kemampuan kerja individu yang
            mencakup aspek pengetahuan, keterampilan/skill, dan sikap kerja yang sesuai dengan
            standar yang ditetapkan;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>3.</span>
          <span
            ><span class="font-bold">Tugas dan tanggung jawab</span> adalah terlampir, merupakan
            bagian yang tidak terpisahkan dari Perjanjian Kerja ini;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>4.</span>
          <span
            ><span class="font-bold">Komitmen</span> untuk bekerja secara profesional sebagaimana
            konsep ataupun sistem yang dibangun Perusahaan, baik secara individu maupun
            kelompok.</span
          >
        </div>
      </div>

      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 3</p>
        <p class="font-bold uppercase">HARI KERJA, JAM KERJA DAN JAM ISTIRAHAT</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            >Hari kerja, jam kerja dan jam istirahat pada dasarnya berlaku sebagaimana kebutuhan
            Perusahaan, namun secara umum berlaku, yaitu :</span
          >
        </div>
        <div class="pl-8 space-y-2">
          <p>
            1) Hari Kerja adalah hari Senin s/d Sabtu, kecuali pada hari-hari tersebut jatuh pada
            hari libur resmi yang ditetapkan oleh Pemerintah Republik Indonesia;
          </p>
          <p>
            2) Jam kerja adalah jam untuk bekerja yang berlaku pada hari kerja mulai jam 07.30 WIB
            s/d 16.30 WIB
          </p>
          <p>
            3) Jam istirahat adalah jam untuk istirahat dan makan siang mulai jam 12.00 WIB s/d
            13.00 WIB (atau dikondisikan sesuai kebutuhan produksi);
          </p>
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            >Perkecualian dari jadwal pada butir (1) pasal ini, yaitu bagi Divisi atau Bagian yang
            sifat pekerjaannya tidak dapat ditinggalkan ataupun terus menerus semisal 24 (dua puluh
            empat) jam sehingga perlu pengaturan khusus seperti sistem regu bergilir 2 (dua) atau 3
            (tiga) shift, selanjutnya jadwal akan diatur oleh masing-masing Divisi dengan tetap
            mengacu kepada ketentuan yang berlaku dan memberikan waktu istirahat secukupnya mulai 30
            - 60 menit;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>3.</span>
          <span
            >Setiap kehadiran dan pulang kerja
            <span class="font-bold uppercase">PIHAK KEDUA</span> wajib mencatatkan kehadirannya
            sesuai dengan jadwal kerja yang telah disetujui.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>4.</span>
          <span
            >Pertukaran dinas dimungkinkan pada kondisi yang mendesak terjadi pada sistem regu
            bergilir, namun demikian hanya dapat dilaksanakan setelah mendapat persetujuan
            Manajer/Penanggungjawab Divisi masing-masing;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>5.</span>
          <span
            >Kepada <span class="font-bold uppercase">PIHAK KEDUA</span> diberikan kesempatan
            secukupnya untuk melaksanakan ibadah wajib .</span
          >
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 4: Pasal 4 -->
    <div class="page-container p-[1cm]">
      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 4</p>
        <p class="font-bold uppercase">GAJI DAN FASILITAS</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            >Atas pelaksanaan pekerjaan oleh <span class="font-bold">PIHAK KEDUA</span> sebagaimana
            tersebut dalam pasal 2 Perjanjian Kerja ini,
            <span class="font-bold">PIHAK PERTAMA</span> akan memberikan Upah dengan komponen
            sebagai berikut :</span
          >
        </div>
        <div class="pl-8 font-bold">
          <p>- TERLAMPIR KETENTUAN GAJI DAN TARGET SALES MOTORIS</p>
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span>Selama masa training 3 bulan akan diberikan upah :</span>
        </div>
        <div class="pl-8 font-bold">
          <p>- TERLAMPIR KETENTUAN GAJI DAN TARGET SALES MOTORIS</p>
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>3.</span>
          <span
            >Para Pihak sepakat bahwa Gaji Pokok sebagaimana dimaksud dalam Perjanjian Kerja ini
            merupakan komponen upah dasar dan bukan merupakan Take Home Pay (THP) yang diterima oleh
            <span class="font-bold">PIHAK KEDUA</span>.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>4.</span>
          <span
            ><span class="font-bold">Take Home Pay (THP)</span> adalah jumlah bersih penghasilan
            yang diterima oleh <span class="font-bold">PIHAK KEDUA</span> setiap bulan, yang
            dihitung dari total penghasilan bruto yang meliputi gaji pokok, tunjangan tetap dan/atau
            tidak tetap, insentif, bonus (apabila ada), setelah dikurangi kewajiban-kewajiban,
            antara lain:</span
          >
        </div>
        <div class="pl-8 space-y-1">
          <p>- Pajak Penghasilan (PPh Pasal 21);</p>
          <p>- Iuran BPJS Kesehatan;</p>
          <p>- Iuran BPJS Ketenagakerjaan;</p>
          <p>
            - Potongan lain yang sah sesuai kebijakan perusahaan dan/atau ketentuan peraturan
            perundang-undangan yang berlaku.
          </p>
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 5: Pasal 4 (cont), 5, 6 -->
    <div class="page-container p-[1cm]">
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>5.</span>
          <span
            ><span class="font-bold">PIHAK KEDUA</span> memahami dan menyetujui bahwa jumlah THP
            yang diterima dapat berbeda setiap bulannya tergantung pada komponen penghasilan dan
            potongan yang berlaku.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>6.</span>
          <span
            ><span class="font-bold">PIHAK KEDUA</span> dengan ini melepaskan hak untuk menuntut
            <span class="font-bold">PIHAMA PERTAMA</span> atas perbedaan antara Gaji Pokok dengan
            THP, sepanjang perhitungan dilakukan sesuai dengan ketentuan dalam Perjanjian Kerja ini
            dan peraturan perundang-undangan yang berlaku.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>7.</span>
          <span
            >Gaji tidak dibayar apabila <span class="font-bold">PIHAK KEDUA</span> tidak melakukan
            pekerjaan;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>8.</span>
          <span>Kerja lembur dan On Call diberikan sesuai dengan ketentuan yang berlaku;</span>
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>9.</span>
          <span
            >Seragam/Pakaian dinas/Pakaian Kerja diberikan dan digunakan sesuai ketentuan yang
            berlaku;</span
          >
        </div>
      </div>

      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 5</p>
        <p class="font-bold uppercase">JAMINAN KESEHATAN DAN PAJAK PENGHASILAN</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            >Jaminan kesehatan diberikan kepada <span class="font-bold">PIHAK KEDUA</span> sesuai
            dengan ketentuan yang berlaku di perusahaan;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            >Pajak yang dikenakan atas gaji atau penghasilan
            <span class="font-bold">PIHAK KEDUA</span> dari
            <span class="font-bold">PIHAK PERTAMA</span> menjadi tanggungan
            <span class="font-bold">PIHAK KEDUA</span> dan akan langsung dipotong setiap bulannya
            sesuai perhitungan ketentuan yang berlaku oleh Divisi Keuangan dan Divisi HR;</span
          >
        </div>
      </div>

      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 6</p>
        <p class="font-bold uppercase">HAK DAN KEWAJIBAN PIHAK PERTAMA</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span><span class="font-bold">PIHAK PERTAMA</span> berhak:</span>
        </div>
        <div class="pl-8 space-y-2">
          <p>
            1) Menuntut atas hasil pekerjaan (prestasi)
            <span class="font-bold">PIHAK KEDUA</span> sebagaimana tugas dan tanggungjawab dalam
            pasal 2 ayat (2) Perjanjian Kerja ini;
          </p>
          <p>
            2) Tidak membayar gaji <span class="font-bold">PIHAK KEDUA</span> sebagaimana mestinya,
            dalam hal <span class="font-bold">PIHAK KEDUA</span> tidak dapat melaksanakan
            kewajibannya;
          </p>
          <p>
            3) Menuntut <span class="font-bold">PIHAK KEDUA</span> untuk mematuhi dan melaksanakan
            secara baik dan bertanggungjawab segala peraturan perusahaan dan/atau ketentuan dalam
            Perjanjian Kerja ini;
          </p>
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 6: Pasal 6 (cont), Pihak 1 Kewajiban, Pasal 7 start -->
    <div class="page-container p-[1cm]">
      <div class="space-y-4 mb-4 text-justify pl-8">
        <p>
          4) Melarang <span class="font-bold">PIHAK KEDUA</span> bekerja rangkap selain ditempat
          <span class="font-bold">PIHAMA PERTAMA</span>, kecuali memperoleh izin tertulis dari
          <span class="font-bold">PIHAK PERTAMA</span>;
        </p>
        <p>
          5) Mencari informasi kepada pihak ketiga mengenai kebenaran keterangan lisan ataupun
          tertulis <span class="font-bold">PIHAK KEDUA</span> yang diberikan kepada
          <span class="font-bold">PIHAK PERTAMA</span> baik selama proses rekrutmen sampai diterima
          sebagai karyawan;
        </p>
        <p>
          6) Memberikan teguran lisan dan/atau peringatan tertulis sampai dengan pemutusan hubungan
          kerja kepada <span class="font-bold">PIHAK KEDUA</span> dalam hal melakukan pelanggaran
          dan/atau tidak mematuhi dan/atau tidak dapat memenuhi Peraturan Perusahaan dan/atau
          Perjanjian Kerja ini, dimana mekanismenya diatur dalam Peraturan Perusahaan/Ketentuan yang
          ada;
        </p>
      </div>

      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span><span class="font-bold">PIHAK PERTAMA</span> berkewajiban :</span>
        </div>
        <div class="pl-8 space-y-2">
          <p>
            1) Membayar gaji <span class="font-bold">PIHAK KEDUA</span> sebagaimana tersebut/diatur
            dalam pasal 4 ayat (1) Perjanjian Kerja ini (atau berdasarkan kesepakatan sebelumnya)
            dan setelah dikurangi kewajiban (misal: Pajak Penghasilan), atas pelaksanaan dan hasil
            pekerjaan (prestasi);
          </p>
          <p>
            2) Memperkerjakan <span class="font-bold">PIHAK KEDUA</span> sebagaimana maksud pasal 2
            Perjanjian Kerja ini atau pada jabatan tertentu sesuai kebutuhan Perusahaan dimasa
            mendatang;
          </p>
          <p>3) Melaksanakan ketentuan yang diatur dalam Perjanjian Kerja ini;</p>
          <p>
            4) Menyediakan sarana-sarana keselamatan kerja dalam usaha mengadakan perlindungan kerja
            sebaik-baiknya yang sesuai dengan ketentuan keselamatan kerja yang berlaku.
          </p>
        </div>
      </div>

      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 7</p>
        <p class="font-bold uppercase">HAK DAN KEWAJIBAN PIHAK KEDUA</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span><span class="font-bold">PIHAK KEDUA</span> berhak atas :</span>
        </div>
        <div class="pl-8 space-y-2">
          <p>
            1) Penerimaan gaji dan/atau tunjangan lainnya sebagaimana hasil pekerjaan (prestasi)
            yang telah dilaksanakan;
          </p>
          <p>
            2) Menerima pekerjaan dalam pasal 2 ayat (1) Perjanjian Kerja ini atau sesuai kebutuhan
            <span class="font-bold">PIHAK PERTAMA</span>;
          </p>
          <p>
            3) Istirahat tahunan atau cuti selama 12 (dua belas) hari kerja sesudah mempunyai masa
            kerja 12 (dua belas) bulan berturut-turut dengan mekanisme pelaksanaan diatur dalam
          </p>
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 7: Pasal 7 (cont), Pihak 2 Kewajiban -->
    <div class="page-container p-[1cm]">
      <div class="space-y-4 mb-4 text-justify pl-8">
        <p>ketentuan Perusahaan;</p>
        <p>
          4) Perlakuan yang layak sesuai dengan peraturan serta ketentuan yang berlaku di
          Perusahaan;
        </p>
        <p>
          5) Perlindungan hukum terhadap ketidakadilan tindakan sewenang-wenang dari atasan atau
          Perusahaan.
        </p>
      </div>

      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span><span class="font-bold">PIHAK KEDUA</span> berkewajiban :</span>
        </div>
        <div class="pl-8 space-y-2">
          <p>
            1) Melaksanakan tugas dan tanggungjawab sebagaimana diatur dalam pasal 2 Perjanjian
            Kerja ini, atau Perintah <span class="font-bold">PIHAK PERTAMA</span> lainnya secara
            baik/benar dan bertanggungjawab (termasuk mengikuti Sosialisasi/Pelatihan/Training).
          </p>
          <p>
            2) Bekerja secara jujur dan rajin dengan dedikasi yang tinggi sesuai kemampuan yang
            dimiliki;
          </p>
          <p>
            3) Mematuhi/menaati ketentuan dalam Perjanjian Kerja ini dan ketentuan lainnya yang
            ditetapkan Perusahaan serta bersedia melaksanakan rotasi atau mutasi sesuai dengan
            kebutuhan Perusahaan, ketentuan lainnya yang ditetapkan Perusahaan berupa Peraturan
            Perusahaan atau Surat Keputusan atau kebijakan lainnya;
          </p>
          <p>
            4) Mematuhi/menaati ketentuan Keselamatan Kerja, Kesehatan Kerja dan Lingkungan yang
            berlaku;
          </p>
          <p>
            5) Menjaga dan memelihara kesehatan, serta senantiasa hidup sehat, baik di tempat kerja
            maupun diluar tempat kerja;
          </p>
          <p>
            6) Menjaga dan menciptakan hubungan industrial yang baik, serta apabila terjadi
            perselisihan hubungan industrial senantiasa akan menyelesaikan secara musyawarah di
            dalam perusahaan.
          </p>
          <p>
            7) Mengembalikan semua alat dan/atau barang-barang Perusahaan serta fasilitas lainnya
            yang dipakai/dipinjamkan/dipinjamnya dalam hal diputuskan hubungan kerjanya dengan
            alasan apapun juga;
          </p>
          <p>
            8) Apabila mendapat seragam/pakaian dinas, memakai sebagaimana peruntukannya dan
            memelihara pakaian kerja/seragam/alat-alat keselamatan pekerja yang disediakan
            <span class="font-bold">PIHAK PERTAMA</span> untuk unit/bagian atau pekerjaan-pekerjaan
            tertentu sesuai dengan ketepatan atau ketentuan yang berlaku;
          </p>
          <p>
            9) Membayar ganti kerugian sebagai atau seluruhnya kepada
            <span class="font-bold">PIHAK PERTAMA</span> yang karena kesengajaan ataupun
            kelalaiannya sehingga terjadi kehilangan atau kerusakan barang milik perusahaan.
          </p>
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 8: Pasal 8 -->
    <div class="page-container p-[1cm]">
      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 8</p>
        <p class="font-bold uppercase">JANGKA WAKTU DAN BERAKHIRNYA PERJANJIAN</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            >Perjanjian kerja ini berlaku terhitung mulai tanggal
            <span class="font-bold underline">{{ formatDate(props.data.startDate) }}</span> sampai
            dengan <span class="font-bold underline">{{ formatDate(props.data.endDate) }}</span
            >, termasuk masa Probation selama <span class="font-bold underline">{{ props.data.probation_months ?? '.....' }}</span> Bulan.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            >Perjanjian kerja ini berlangsung terus sampai saat berakhirnya waktu yang telah
            ditetapkan dalam ayat (1) diatas, kecuali karena :</span
          >
        </div>
        <div class="pl-8 space-y-2">
          <div class="grid grid-cols-[30px_1fr]">
            <span>1)</span>
            <span
              >Ketidaksesuaian salah satu pihak selama berlangsungnya Perjanjian Kerja ini, sehingga
              <span class="font-bold underline"
                >masing-masing pihak dimungkinkan untuk mengakhiri sebelum jangka waktu berakhirnya
                PKWT ini dengan pemberitahuan paling lambat 7 Hari sebelumnya</span
              >, disini pertimbangan <span class="font-bold">PIHAK PERTAMA</span> dalam melakukan
              penilaian terhadap <span class="font-bold">PIHAK KEDUA</span> yaitu ketidakmampuan
              melaksanakan ketentuan Pasal 2 Perjanjian Kerja ini baik ayat 1, ayat 2, ayat 3 dan
              ayat 4;</span
            >
          </div>
          <div class="grid grid-cols-[30px_1fr]">
            <span>2)</span>
            <span
              >Kesalahan berat yang dilakukan <span class="font-bold">PIHAK KEDUA</span>, antara
              lain :</span
            >
          </div>
          <div class="pl-8 grid grid-cols-[30px_1fr] gap-y-1">
            <span>a.</span
            ><span>Menyimpang dari Standar Mutu Produksi dan/atau Standar Profesi;</span>
            <span>b.</span
            ><span
              >Bekerja rangkap tanpa sepengetahuan dan persetujuan
              <span class="font-bold">PIHAK PERTAMA</span>;</span
            >
            <span>c.</span
            ><span
              >Memberikan keterangan palsu atau dipalsukan baik pada saat proses rekrutment sampai
              dengan menjadi karyawan;</span
            >
            <span>d.</span
            ><span>Mabuk, memakai obat-obatan terlarang atau narkotika di tempat kerja;</span>
            <span>e.</span
            ><span
              >Mencuri, menggelapkan, menipu atau melakukan kejahatan lainnya yang melanggar
              ketentuan Hukum yang berlaku di Negara Kesatuan Republik Indonesia;</span
            >
            <span>f.</span
            ><span
              >Menganiaya, menghina, ataupun mengancam
              <span class="font-bold">PIHAK PERTAMA</span> atau teman sekerja atau
              manajemen/pelanggan {{ companyName }};</span
            >
            <span>g.</span><span>Melakukan sesuatu yang bertentangan dengan hukum;</span>
            <span>h.</span
            ><span>Berbuat asusila baik ditempat kerja maupun diluar tempat kerja;</span>
            <span>i.</span
            ><span
              >Dengan sengaja atau kecerobohanya/lalai mengakibatkan barang milik atau yang disewa
              Perusahaan Hilang atau rusak atau dalam keadaan berbahaya/membahayakan;</span
            >
            <span>j.</span
            ><span
              >Tidak masuk kerja tanpa pemberitahuan 5 (lima) hari kerja secara berturut-turut atau
              tidak berturut-turut selama perjanjian kerja ini berlangsung;</span
            >
            <span>k.</span><span>Membongkar rahasia perusahaan yang seharusnya di rahasiakan;</span>
            <span>l.</span
            ><span
              >Setelah peringatan tertulis terakhir, tetap melakukan pelanggaran yang sama;</span
            >
            <span>m.</span
            ><span
              >Dilarang menerima gratifikasi (Uang, Hadiah, Parsel) dari pihak manapun tanpa
              ketentuan Perusahaan;</span
            >
          </div>
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 9: Pasal 9, 10 & Closing -->
    <div class="page-container p-[1cm]">
      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 9</p>
        <p class="font-bold uppercase">PENYELESAIAN MASALAH</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            >Apabila terjadi perselisihan dalam perjanjian kerjasama ini, maka akan diselesaikan
            secara musyawarah untuk mufakat atau melalui mediasi oleh Divisi HR</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            >Apabila penyelesaian secara musyawarah tidak mencapai kata mufakat, maka
            <span class="font-bold">PARA PIHAK</span> sepakat untuk menyelesaikan perselisihan ini
            melalui kepaniteraan Pengadilan Negeri Setempat (dimana PKWT ini di tandatangani).</span
          >
        </div>
      </div>

      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 10</p>
        <p class="font-bold uppercase">PENUTUP</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            >Agar <span class="font-bold">PIHAK KEDUA</span> memahami maksud dan isi dari Perjanjian
            Kerja ini maka <span class="font-bold">PIHAK KEDUA</span> sebelum menandatangani agar
            membaca Perjanjian Kerja ini secara cermat dan seksama;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            >Peraturan Perusahaan dan ketentuan lainnya yang telah ditetapkan atau akan ditetapkan
            atau yang akan ditetapkan, merupakan bagian yang tidak terpisahkan dari Perjanjian Kerja
            ini;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>3.</span>
          <span
            >Untuk hal-hal lain yang tidak cukup ataupun belum diatur dalam perjanjian kerja ini,
            maka akan diatur dalam suatu perjanjian tambahan (<span class="italic">Addendum</span>)
            dan/atau suatu penambahan lampiran yang akan dituangkan secara tertulis serta
            ditandatangani oleh <span class="font-bold">PARA PIHAK</span>, Dimana perjanjian
            tambahan dan/atau penambahan lampiran tersebut adalah menjadi satu kesatuan yang tidak
            terpisahkan dengan perjanjian ini;</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>4.</span>
          <span
            ><span class="font-bold">PARA PIHAK</span> menjamin kebenaran atas identitas diri dan
            kewenangan untuk menandatangani perjanjian ini dan mengetahui akibat hukum yang timbul
            dari perjanjian ini serta mengikat bagi <span class="font-bold">PARA PIHAK</span>.</span
          >
        </div>
      </div>

      <p class="mb-4 text-justify">
        Demikian Perjanjian Kerja ini dibuat dengan sesungguhnya dan tidak ada paksaan, yang mana
        sebelum menandatangani perjanjian ini <span class="font-bold">PARA PIHAK</span> telah
        membaca, memahami dan menyetujui seluruh isi dari perjanjian kerjasama ini yang memuat
        sebagian norma dan syarat kerja, serta hak dan kewajiban
        <span class="font-bold">PARA PIHAK</span>, sehingga untuk dijunjung tinggi dan dilaksanakan
        secara baik dengan penuh tanggungjawab, yang ditandatangani oleh
        <span class="font-bold">PARA PIHAK</span> dan dibuat rangkap 2 (dua) masing-masing mempunyai
        kekuatan hukum yang sama.
      </p>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 10: Signatures Main -->
    <div class="page-container p-[1cm]">
      <div class="grid grid-cols-2 border border-black mt-10 min-h-[300px]">
        <div class="border-r border-black p-4 text-center flex flex-col items-center">
          <p class="font-bold uppercase mb-4">PIHAK PERTAMA</p>
          <div class="h-24 my-2"></div>
          <div class="mt-auto">
            <p class="font-bold underline uppercase">{{ props.data.representativeName }}</p>
            <p class="font-bold uppercase">{{ props.data.representativePosition }}</p>
          </div>
        </div>
        <div class="p-4 text-center flex flex-col items-center">
          <p class="font-bold uppercase mb-4">PIHAK KEDUA</p>
          <div v-if="props.data.employeeSignatureUrl" class="my-2 h-24 flex items-center justify-center">
            <img :src="getSignatureFullUrl(props.data.employeeSignatureUrl)" alt="Tanda Tangan Karyawan" class="h-24 object-contain" />
          </div>
          <div v-else class="h-24 my-2"></div>
          <div class="mt-auto">
            <p class="font-bold underline uppercase">
              {{ props.data.employee?.first_name }} {{ props.data.employee?.last_name || '' }}
            </p>
            <p class="font-bold uppercase">KARYAWAN</p>
          </div>
        </div>
        <div
          class="col-span-2 border-t border-black p-4 text-center flex flex-col items-center min-h-[150px]"
        >
          <p class="italic mb-16">Mengetahui,</p>
          <div class="mt-auto">
            <div class="w-48 border-b border-black mb-1"></div>
            <p class="font-bold uppercase">GM/MANAGER MARKETING</p>
          </div>
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 11: Addendum Header & Pasal 1 -->
    <div class="page-container p-[1cm]">
      <div class="text-center mb-4 space-y-1">
        <h1 class="font-bold underline uppercase text-[12pt]">
          ADDENDUM PERJANJIAN KERJA WAKTU TERTENTU (PKWT)
        </h1>
        <p class="font-bold">
          NOMOR : 015/DEMO-ADD-PKWT/{{ getYear(props.data.agreementDate) }}
        </p>
      </div>

      <p class="mb-3 text-justify">
        Addendum ini merupakan bagian yang tidak terpisahkan dari PKWT NOMOR
        {{ props.data.documentNumber }}/DEMO-PKWT/{{
          getYear(props.data.agreementDate)
        }}
        tanggal {{ formatDate(props.data.agreementDate) }} yang dibuat antara:
      </p>

      <!-- Pihak Pertama Addendum -->
      <div class="grid grid-cols-[30px_120px_20px_1fr] mb-3 gap-y-1">
        <span>(1)</span><span>Nama</span><span>:</span
        ><span class="font-bold">{{ companyName }}</span> <span></span><span>Alamat</span
        ><span>:</span>
        <div class="grid grid-cols-[160px_20px_1fr]">
          <span>Alamat Perusahaan</span><span>:</span><span>{{ companyAddress || '....................' }}</span>
        </div>
        <div class="col-span-4 mt-2 pl-[30px]">
          <p class="text-justify">
            Dalam hal ini diwakili oleh
            <span class="font-bold">{{ props.data.representativeName }}</span> selaku
            <span class="font-bold">{{ props.data.representativePosition }}</span> di {{ companyName }} dan menurut jabatannya bertindak untuk dan atas nama
            {{ companyName }} sebagai perusahaan pemberi kerja, oleh dan karenanya sah dan berwenang
            untuk bertindak untuk dan atas nama perusahaan (selanjutnya dalam perjanjian ini disebut "<span
              class="font-bold"
              >PIHAK PERTAMA</span
            >").
          </p>
        </div>
      </div>

      <!-- Pihak Kedua Addendum -->
      <div class="grid grid-cols-[30px_120px_20px_1fr] mb-3 gap-y-1">
        <span>(2)</span><span>Nama</span><span>:</span
        ><span class="font-bold"
          >{{ props.data.employee?.first_name }} {{ props.data.employee?.last_name }}</span
        >
        <span></span><span>Tempat/Tanggal Lahir</span><span>:</span
        ><span class="font-bold"
          >{{ props.data.employee?.birth_place }},
          {{ formatDate(props.data.employee?.birth_date) }}</span
        >
        <span></span><span>Jenis Kelamin</span><span>:</span
        ><span class="font-bold">{{ props.data.employee?.gender }}</span> <span></span
        ><span>Alamat</span><span>:</span>
        <div class="grid grid-cols-[160px_20px_1fr]">
          <span>Nama/Nomor Jalan</span><span>:</span
          ><span>{{ props.data.employee?.address_ktp || '....................' }}</span>
          <span>RT/RW</span><span>:</span><span>....................</span> <span>Kelurahan</span
          ><span>:</span><span>....................</span> <span>Kecamatan</span><span>:</span
          ><span>....................</span> <span>Kabupaten/Kota</span><span>:</span
          ><span>....................</span> <span>Provinsi</span><span>:</span
          ><span>....................</span>
        </div>
        <span></span><span>NIK</span><span>:</span
        ><span class="font-bold">{{
          props.data.employee?.identity_number || '....................'
        }}</span>
        <div class="col-span-4 mt-2 pl-[30px]">
          <p class="text-justify">
            (selanjutnya dalam perjanjian ini disebut "<span class="font-bold">PIHAK KEDUA</span>"
            yang merupakan Penerima Kerja/Karyawan).
          </p>
        </div>
      </div>

      <div class="text-center mb-4 mt-8">
        <p class="font-bold uppercase underline">PASAL 1</p>
        <p class="font-bold uppercase">MAKSUD DAN TUJUAN</p>
      </div>
      <p class="mb-4 text-justify">
        Addendum ini dibuat sebagai penambahan ketentuan dalam PKWT NOMOR
        {{ props.data.documentNumber }}/DEMO-PKWT/{{ getYear(props.data.agreementDate) }} yang
        sebelumnya telah disepakati oleh <span class="font-bold">PARA PIHAK</span>.
      </p>

      <div class="mt-auto flex justify-end pt-8">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>

    <!-- PAGE 12: Addendum Pasal 2 & Signatures -->
    <div class="page-container p-[1cm]">
      <div class="text-center mb-4">
        <p class="font-bold uppercase underline">PASAL 2</p>
        <p class="font-bold uppercase">KETENTUAN TIDAK MENUNTUT KOMPENSASI</p>
      </div>
      <div class="space-y-4 mb-4 text-justify">
        <div class="grid grid-cols-[30px_1fr]">
          <span>1.</span>
          <span
            ><span class="font-bold">PIHAK KEDUA</span> memahami dan menyetujui bahwa selama masa
            Perjanjian Kerja Waktu Tertentu (PKWT) berlangsung,
            <span class="font-bold">PIHAK PERTAMA</span>
            <span class="font-bold underline"
              >tidak memberlakukan denda kepada PIHAK KEDUA apabila mengundurkan diri sebelum masa
              kontrak berakhir</span
            >.</span
          >
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>2.</span>
          <span
            >Sehubungan dengan ketentuan tersebut, <span class="font-bold">PIHAK KEDUA</span> dengan
            ini menyatakan
            <span class="font-bold underline"
              >tidak akan menuntut kompensasi, ganti rugi, atau pembayaran apapun</span
            >
            kepada <span class="font-bold">PIHAK PERTAMA</span> apabila:</span
          >
        </div>
        <div class="pl-8 space-y-2">
          <p>
            A. <span class="font-bold">PIHAK KEDUA</span> mengundurkan diri sebelum masa kontrak
            berakhir, atau
          </p>
          <p>
            B. Hubungan kerja diakhiri berdasarkan kebijakan Perusahaan/Manajemen sesuai ketentuan
            yang berlaku, pengakhiran hubungan kerja sewaktu waktu dapat dilakukan dengan alasan
            sebagai berikut :
          </p>
          <div class="pl-8 space-y-1">
            <p>
              - Tidak tercapainya target penjualan atau KPI
              <span class="font-bold">PIHAK KEDUA</span>
            </p>
            <p>
              - Dan sesuai Pasal 8 Ayat 2 Perjanjian Kerja Waktu Tertentu
              <span class="font-bold">PIHAK KEDUA</span>
            </p>
          </div>
        </div>
        <div class="grid grid-cols-[30px_1fr]">
          <span>3.</span>
          <span
            ><span class="font-bold">PIHAK KEDUA</span> menyatakan bahwa keputusan ini dipahami dan
            disetujui secara sadar tanpa adanya paksaan dari pihak manapun.</span
          >
        </div>
      </div>

      <div class="grid grid-cols-2 border border-black mt-2 min-h-[300px]">
        <div class="border-r border-black p-4 text-center flex flex-col items-center">
          <p class="font-bold uppercase mb-4">PIHAK PERTAMA</p>
          <div v-if="props.data.hr_signature_url" class="my-2 h-24 flex items-center justify-center">
            <img :src="getSignatureFullUrl(props.data.hr_signature_url)" alt="Tanda Tangan HRD" class="h-24 object-contain" />
          </div>
          <div v-else class="h-24 my-2"></div>
          <div class="mt-auto">
            <p class="font-bold underline uppercase">{{ props.data.representativeName }}</p>
            <p class="font-bold uppercase">{{ props.data.representativePosition }}</p>
          </div>
        </div>
        <div class="p-4 text-center flex flex-col items-center">
          <p class="font-bold uppercase mb-4">PIHAK KEDUA</p>
          <div v-if="props.data.employeeSignatureUrl" class="my-2 h-24 flex items-center justify-center">
            <img :src="getSignatureFullUrl(props.data.employeeSignatureUrl)" alt="Tanda Tangan Karyawan" class="h-24 object-contain" />
          </div>
          <div v-else class="h-24 my-2"></div>
          <div class="mt-auto">
            <p class="font-bold underline uppercase">
              {{ props.data.employee?.first_name }} {{ props.data.employee?.last_name || '' }}
            </p>
            <p class="font-bold uppercase">KARYAWAN</p>
          </div>
        </div>
        <div
          class="col-span-2 border-t border-black p-4 text-center flex flex-col items-center min-h-[150px]"
        >
          <p class="italic mb-16">Mengetahui,</p>
          <div class="mt-auto">
            <div class="w-48 border-b border-black mb-1"></div>
            <p class="font-bold uppercase">GM/MANAGER MARKETING</p>
          </div>
        </div>
      </div>

      <div class="mt-auto flex justify-end pt-12">
        <table class="border-collapse border border-black text-[8pt] w-[180px]">
          <tr>
            <td colspan="2" class="border border-black text-center font-bold py-1 uppercase">
              PARAF
            </td>
          </tr>
          <tr class="h-8">
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 1</td>
            <td class="border border-black text-center w-1/2 align-top py-0.5">PIHAK 2</td>
          </tr>
          <tr class="h-10">
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.hr_signature_url"
                :src="getSignatureFullUrl(props.data.hr_signature_url)"
                alt="Paraf HRD"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
            <td class="border border-black w-1/2 text-center p-0.5">
              <img
                v-if="props.data.employeeSignatureUrl"
                :src="getSignatureFullUrl(props.data.employeeSignatureUrl)"
                alt="Paraf Karyawan"
                class="h-8 max-w-full mx-auto object-contain"
              />
            </td>
          </tr>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
@media print {
  @page {
    size: A4;
    margin: 0;
  }
  body {
    background: white;
  }
  .pkwt-container {
    width: 100%;
    max-width: none;
    box-shadow: none;
    margin: 0;
    padding: 0;
  }
  .page-container {
    min-height: 26cm;
    page-break-after: always;
    display: flex;
    flex-direction: column;
  }
}

.pkwt-container {
  font-family: 'Times New Roman', Times, serif;
}

.page-container {
  min-height: 26cm;
  display: flex;
  flex-direction: column;
}
</style>
