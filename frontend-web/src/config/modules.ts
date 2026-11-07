const validModules = new Set(['hris', 'factory', 'warehouse', 'purchasing', 'sales', 'marketplace', 'branding', 'finance', 'rajatransport'])
const configured = (import.meta.env.VITE_ENABLED_MODULES || 'hris,factory,warehouse,purchasing,sales,marketplace,branding,finance,rajatransport')
  .split(',').map((value: string) => value.trim().toLowerCase()).filter(Boolean)
export const enabledModules = new Set(configured.filter((value: string) => validModules.has(value)))
enabledModules.add('hris')
export const isModuleEnabled = (module: string): boolean => enabledModules.has(module)
export const isPkwtEnabled = import.meta.env.VITE_ENABLE_PKWT === 'true'

export const moduleForPath = (path: string): string => {
  if (path.startsWith('/vehicles')) return 'factory'
  if (path.startsWith('/branding-integration')) return 'branding'
  if (path.startsWith('/purchasing')) return 'purchasing'
  if (path.startsWith('/factory')) return 'factory'
  if (path.startsWith('/warehouse')) return 'warehouse'
  if (path.startsWith('/finance')) return 'finance'
  if (path.startsWith('/marketplace')) return 'marketplace'
  if (path.startsWith('/rajatransport')) return 'rajatransport'
  if (path.startsWith('/sales/spanduk') || path.startsWith('/sales/branding')) return 'branding'
  if (path.startsWith('/sales')) return 'sales'
  return 'hris'
}
