const readSingleQueryValue = (value: unknown): string => {
  if (typeof value === 'string') return value
  if (Array.isArray(value) && typeof value[0] === 'string') return value[0]
  return ''
}

export const pickDateRangeQuery = (query: Record<string, unknown>) => {
  const startTime = readSingleQueryValue(query.start_time)
  const endTime = readSingleQueryValue(query.end_time)
  const startDate = readSingleQueryValue(query.start_date)
  const endDate = readSingleQueryValue(query.end_date)
  return {
    startTime: startTime || startDate || '',
    endTime: endTime || endDate || ''
  }
}

export const appendDateRangeQuery = (target: Record<string, string>, query: Record<string, unknown>) => {
  const { startTime, endTime } = pickDateRangeQuery(query)
  if (startTime) target.start_time = startTime
  if (endTime) target.end_time = endTime
}
