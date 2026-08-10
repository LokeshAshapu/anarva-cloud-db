'use client'

import React from 'react'
import { CloudStatus } from './CloudStatus'
import { ResourceStatus } from '@/types/resource'

export interface CloudResourceStatusProps {
  status: ResourceStatus | string
}

export function CloudResourceStatus({ status }: CloudResourceStatusProps) {
  return <CloudStatus status={status} />
}
