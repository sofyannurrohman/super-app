import axios from 'axios'

const baseUrl = 'https://www.emsifa.com/api-wilayah-indonesia/api'

export const regionApi = {
  getProvinces: async () => {
    const res = await axios.get(`${baseUrl}/provinces.json`)
    return res.data
  },
  getRegencies: async (provinceId: string) => {
    const res = await axios.get(`${baseUrl}/regencies/${provinceId}.json`)
    return res.data
  },
  getDistricts: async (regencyId: string) => {
    const res = await axios.get(`${baseUrl}/districts/${regencyId}.json`)
    return res.data
  },
  getVillages: async (districtId: string) => {
    const res = await axios.get(`${baseUrl}/villages/${districtId}.json`)
    return res.data
  }
}
