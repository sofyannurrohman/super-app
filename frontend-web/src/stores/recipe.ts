import { defineStore } from 'pinia'
import apiClient from '@/api/axios'

export interface RecipeItem {
  id?: string
  raw_product_id: string
  quantity: number
  raw_product?: any
}

export interface Recipe {
  id?: string
  finished_product_id: string
  description?: string
  items: RecipeItem[]
  finished_product?: any
}

export const useRecipeStore = defineStore('recipe', {
  state: () => ({
    recipes: [] as Recipe[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchRecipes() {
      this.loading = true
      try {
        const res = await apiClient.get('/admin/factory/recipes')
        this.recipes = res.data || []
      } catch (err: any) {
        this.error = 'Gagal memuat resep'
      } finally {
        this.loading = false
      }
    },

    async fetchRecipeByProduct(productID: string) {
      try {
        const res = await apiClient.get(`/admin/factory/recipes/${productID}`)
        return res.data
      } catch (err) {
        return null
      }
    },

    async createRecipe(data: Recipe) {
      await apiClient.post('/admin/factory/recipes', data)
      await this.fetchRecipes()
    },

    async deleteRecipe(id: string) {
      await apiClient.delete(`/admin/factory/recipes/${id}`)
      await this.fetchRecipes()
    },
  },
})
