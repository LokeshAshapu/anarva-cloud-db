// Anarva Resource Identifier (ARNV) Generator

export function generateARNV(
  resourceType: string,
  regionId: string,
  projectId: string,
  resourceName: string
): string {
  let typeSlug = resourceType.toLowerCase()
  switch (typeSlug) {
    case 'database':
      typeSlug = 'db'
      break
    case 'storage_bucket':
    case 'storage':
      typeSlug = 's3'
      break
    case 'compute':
      typeSlug = 'vm'
      break
    case 'network':
      typeSlug = 'vpc'
      break
    case 'backup':
      typeSlug = 'bak'
      break
    case 'replica':
      typeSlug = 'rep'
      break
  }

  const cleanName = resourceName.toLowerCase().replace(/\s+/g, '-')
  return `arnv:${typeSlug}:${regionId}:${projectId}:${resourceType.toLowerCase()}/${cleanName}`
}
