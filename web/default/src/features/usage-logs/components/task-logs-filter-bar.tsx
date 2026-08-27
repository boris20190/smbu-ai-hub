import { useState, useCallback } from 'react'
import { useQueryClient, useIsFetching } from '@tanstack/react-query'
import { useNavigate, getRouteApi } from '@tanstack/react-router'
import { type Table } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'
import { useIsAdmin } from '@/hooks/use-admin'
import { Input } from '@/components/ui/input'
import { DataTableToolbar } from '@/components/data-table'
import { buildSearchParams } from '../lib/filter'
import { getDefaultTimeRange } from '../lib/utils'
import type { DrawingLogFilters, LogCategory, TaskLogFilters } from '../types'
import { CompactDateTimeRangePicker } from './compact-date-time-range-picker'

const route = getRouteApi('/_authenticated/usage-logs/$section')

type TaskLikeLogCategory = Extract<LogCategory, 'drawing' | 'task'>
type TaskLogsFilters = DrawingLogFilters | TaskLogFilters
type TaskLogsSearchParams = ReturnType<typeof route.useSearch>

interface TaskLogsFilterBarProps<TData> {
  table: Table<TData>
  logCategory: TaskLikeLogCategory
}

function getFilterValue(
  filters: TaskLogsFilters,
  logCategory: TaskLikeLogCategory
): string {
  if (logCategory === 'drawing') {
    return (filters as DrawingLogFilters).mjId || ''
  }
  return (filters as TaskLogFilters).taskId || ''
}

function setFilterValue(
  filters: TaskLogsFilters,
  logCategory: TaskLikeLogCategory,
  value: string
): TaskLogsFilters {
  if (logCategory === 'drawing') {
    return { ...filters, mjId: value }
  }
  return { ...filters, taskId: value }
}

function getSearchParamKey(value: unknown): string {
  if (Array.isArray(value)) return value.map(String).join(',')
  return value == null ? '' : String(value)
}

function getTaskLogsSearchKey(
  searchParams: TaskLogsSearchParams,
  logCategory: TaskLikeLogCategory
): string {
  return [
    logCategory,
    searchParams.startTime,
    searchParams.endTime,
    searchParams.channel,
    searchParams.filter,
  ]
    .map(getSearchParamKey)
    .join('|')
}

function buildInitialTaskLogsFilters(
  searchParams: TaskLogsSearchParams,
  logCategory: TaskLikeLogCategory
): TaskLogsFilters {
  const { start, end } = getDefaultTimeRange()
  const baseFilters = {
    startTime: searchParams.startTime
      ? new Date(searchParams.startTime)
      : start,
    endTime: searchParams.endTime ? new Date(searchParams.endTime) : end,
    ...(searchParams.channel ? { channel: String(searchParams.channel) } : {}),
  }

  if (logCategory === 'drawing') {
    return {
      ...baseFilters,
      ...(searchParams.filter ? { mjId: searchParams.filter } : {}),
    }
  }

  return {
    ...baseFilters,
    ...(searchParams.filter ? { taskId: searchParams.filter } : {}),
  }
}

export function TaskLogsFilterBar<TData>(props: TaskLogsFilterBarProps<TData>) {
  const searchParams = route.useSearch()
  const searchKey = getTaskLogsSearchKey(searchParams, props.logCategory)

  return (
    <TaskLogsFilterBarContent
      key={searchKey}
      {...props}
      searchParams={searchParams}
    />
  )
}

interface TaskLogsFilterBarContentProps<
  TData,
> extends TaskLogsFilterBarProps<TData> {
  searchParams: TaskLogsSearchParams
}

function TaskLogsFilterBarContent<TData>(
  props: TaskLogsFilterBarContentProps<TData>
) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const isAdmin = useIsAdmin()
  const fetchingLogs = useIsFetching({ queryKey: ['logs'] })

  const [filters, setFilters] = useState<TaskLogsFilters>(() => {
    return buildInitialTaskLogsFilters(props.searchParams, props.logCategory)
  })

  const handleChange = useCallback(
    (field: keyof TaskLogsFilters, value: Date | string | undefined) => {
      setFilters((prev) => ({ ...prev, [field]: value }))
    },
    []
  )

  const handleApply = useCallback(() => {
    const filterParams = buildSearchParams(filters, props.logCategory)
    navigate({
      to: '/usage-logs/$section',
      params: { section: props.logCategory },
      search: {
        ...filterParams,
        page: 1,
      },
    })
    queryClient.invalidateQueries({ queryKey: ['logs'] })
  }, [filters, navigate, props.logCategory, queryClient])

  const handleReset = useCallback(() => {
    const { start, end } = getDefaultTimeRange()
    const resetFilters: TaskLogsFilters = { startTime: start, endTime: end }
    setFilters(resetFilters)

    navigate({
      to: '/usage-logs/$section',
      params: { section: props.logCategory },
      search: {
        page: 1,
        startTime: start.getTime(),
        endTime: end.getTime(),
      },
    })
    queryClient.invalidateQueries({ queryKey: ['logs'] })
  }, [navigate, props.logCategory, queryClient])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') handleApply()
    },
    [handleApply]
  )

  const handleFilterChange = useCallback(
    (value: string) => {
      setFilters((prev) => setFilterValue(prev, props.logCategory, value))
    },
    [props.logCategory]
  )

  const filterValue = getFilterValue(filters, props.logCategory)
  const placeholder =
    props.logCategory === 'drawing'
      ? t('Filter by Midjourney task ID')
      : t('Filter by task ID')
  const inputClass = 'w-full sm:w-[180px] lg:w-[200px]'
  const hasAdditionalFilters = !!filterValue || !!filters.channel

  return (
    <DataTableToolbar
      table={props.table}
      customSearch={
        <CompactDateTimeRangePicker
          start={filters.startTime}
          end={filters.endTime}
          onChange={({ start, end }) => {
            handleChange('startTime', start)
            handleChange('endTime', end)
          }}
          className='w-full sm:w-[340px]'
        />
      }
      additionalSearch={
        <>
          <Input
            aria-label={t('Task ID')}
            placeholder={placeholder}
            value={filterValue}
            onChange={(e) => handleFilterChange(e.target.value)}
            onKeyDown={handleKeyDown}
            className={inputClass}
          />
          {isAdmin && (
            <Input
              placeholder={t('Channel ID')}
              value={filters.channel || ''}
              onChange={(e) => handleChange('channel', e.target.value)}
              onKeyDown={handleKeyDown}
              className={inputClass}
            />
          )}
        </>
      }
      hasAdditionalFilters={hasAdditionalFilters}
      onSearch={handleApply}
      searchLoading={fetchingLogs > 0}
      onReset={handleReset}
    />
  )
}
