import React from 'react';
import BaseLayout from '@/components/layout/layout';


export default function Layout({ children }: { children: React.ReactNode }) {

  return (
    <BaseLayout>
      {children}
    </BaseLayout>
  );
}