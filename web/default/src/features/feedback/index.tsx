import { useMemo, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Send, MessageSquareText, Inbox } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { formatTimestampToDate } from '@/lib/format'
import { useIsAdmin } from '@/hooks/use-admin'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from '@/components/ui/empty'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { SectionPageLayout } from '@/components/layout'
import { getFeedbacks, submitFeedback } from './api'

type FeedbackFormValues = {
  title: string
  content: string
  contact: string
}

const FEEDBACK_PAGE_SIZE = 10

export function Feedback() {
  const { t } = useTranslation()
  const isAdmin = useIsAdmin()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)

  const schema = useMemo(
    () =>
      z.object({
        title: z
          .string()
          .trim()
          .min(2, t('Feedback title must be 2-80 characters.'))
          .max(80, t('Feedback title must be 2-80 characters.')),
        content: z
          .string()
          .trim()
          .min(5, t('Feedback content must be 5-2000 characters.'))
          .max(2000, t('Feedback content must be 5-2000 characters.')),
        contact: z
          .string()
          .trim()
          .max(120, t('Contact information must be within 120 characters.')),
      }),
    [t]
  )

  const form = useForm<FeedbackFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: '',
      content: '',
      contact: '',
    },
  })

  const feedbackQuery = useQuery({
    queryKey: ['feedback', page, FEEDBACK_PAGE_SIZE],
    enabled: isAdmin,
    queryFn: async () => {
      const res = await getFeedbacks({ p: page, page_size: FEEDBACK_PAGE_SIZE })
      if (!res.success || !res.data) {
        throw new Error(res.message || 'Failed to load feedback')
      }
      return res.data
    },
  })

  const submitMutation = useMutation({
    mutationFn: async (values: FeedbackFormValues) => {
      const res = await submitFeedback({
        title: values.title,
        content: values.content,
        contact: values.contact || undefined,
      })
      if (!res.success) {
        throw new Error(res.message || t('Feedback submission failed'))
      }
      return res.data
    },
    onSuccess: () => {
      toast.success(t('Feedback submitted'))
      form.reset()
      if (isAdmin) {
        void queryClient.invalidateQueries({ queryKey: ['feedback'] })
      }
    },
    onError: (error) => {
      toast.error(
        error instanceof Error ? error.message : t('Feedback submission failed')
      )
    },
  })

  const feedbackData = feedbackQuery.data
  const total = feedbackData?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / FEEDBACK_PAGE_SIZE))
  const items = feedbackData?.items ?? []

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Feedback')}</SectionPageLayout.Title>
      <SectionPageLayout.Description>
        {t('Submit platform feedback. Administrators can review it.')}
      </SectionPageLayout.Description>
      <SectionPageLayout.Content>
        <div className='grid gap-4 xl:grid-cols-[minmax(360px,0.85fr)_minmax(0,1.15fr)]'>
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2'>
                <MessageSquareText className='size-4' />
                {t('Submit feedback')}
              </CardTitle>
              <CardDescription>
                {t('Tell us about issues, requests, or suggestions.')}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form
                  className='space-y-4'
                  onSubmit={form.handleSubmit((values) =>
                    submitMutation.mutate(values)
                  )}
                >
                  <FormField
                    control={form.control}
                    name='title'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('Feedback title')}</FormLabel>
                        <FormControl>
                          <Input
                            placeholder={t('Briefly describe the feedback')}
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name='content'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('Feedback content')}</FormLabel>
                        <FormControl>
                          <Textarea
                            className='min-h-36 resize-y'
                            placeholder={t(
                              'Describe the issue, request, or suggestion'
                            )}
                            {...field}
                          />
                        </FormControl>
                        <FormDescription>
                          {t('Please avoid sharing API keys or passwords.')}
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name='contact'
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t('Contact information')}</FormLabel>
                        <FormControl>
                          <Input
                            placeholder={t('Optional contact details')}
                            {...field}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <Button type='submit' disabled={submitMutation.isPending}>
                    <Send />
                    {submitMutation.isPending
                      ? t('Submitting...')
                      : t('Submit feedback')}
                  </Button>
                </form>
              </Form>
            </CardContent>
          </Card>

          {isAdmin && (
            <Card>
              <CardHeader>
                <CardTitle className='flex items-center gap-2'>
                  <Inbox className='size-4' />
                  {t('Feedback inbox')}
                </CardTitle>
                <CardDescription>
                  {t('Feedback submitted by all users.')}
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-4'>
                {items.length > 0 ? (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('Title')}</TableHead>
                        <TableHead>{t('User')}</TableHead>
                        <TableHead>{t('Contact')}</TableHead>
                        <TableHead>{t('Submitted at')}</TableHead>
                        <TableHead className='min-w-72'>
                          {t('Content')}
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {items.map((item) => (
                        <TableRow key={item.id}>
                          <TableCell className='max-w-56 font-medium whitespace-normal'>
                            {item.title}
                          </TableCell>
                          <TableCell>{item.username || '-'}</TableCell>
                          <TableCell className='max-w-56 whitespace-normal'>
                            {item.contact || '-'}
                          </TableCell>
                          <TableCell>
                            {formatTimestampToDate(item.created_time)}
                          </TableCell>
                          <TableCell className='max-w-xl whitespace-pre-wrap'>
                            {item.content}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                ) : (
                  <Empty className='min-h-60 border'>
                    <EmptyHeader>
                      <EmptyMedia variant='icon'>
                        <Inbox />
                      </EmptyMedia>
                      <EmptyTitle>
                        {feedbackQuery.isLoading
                          ? t('Loading')
                          : t('No feedback yet')}
                      </EmptyTitle>
                      <EmptyDescription>
                        {t('Submitted feedback will appear here.')}
                      </EmptyDescription>
                    </EmptyHeader>
                  </Empty>
                )}

                <div className='flex flex-wrap items-center justify-between gap-3 text-sm'>
                  <span className='text-muted-foreground'>
                    {t('Total')}: {total}
                  </span>
                  <div className='flex items-center gap-2'>
                    <Button
                      type='button'
                      variant='outline'
                      size='sm'
                      disabled={page <= 1 || feedbackQuery.isFetching}
                      onClick={() =>
                        setPage((current) => Math.max(1, current - 1))
                      }
                    >
                      {t('Previous')}
                    </Button>
                    <span className='text-muted-foreground min-w-20 text-center'>
                      {page} / {totalPages}
                    </span>
                    <Button
                      type='button'
                      variant='outline'
                      size='sm'
                      disabled={page >= totalPages || feedbackQuery.isFetching}
                      onClick={() =>
                        setPage((current) => Math.min(totalPages, current + 1))
                      }
                    >
                      {t('Next')}
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
