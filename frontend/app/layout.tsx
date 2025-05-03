import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import NextTopLoader from 'nextjs-toploader';
import { Toaster } from 'sonner';
import { WebSocketProvider } from "@/lib/context/web-socket-context";
import { QueryProvider } from "@/lib/context/query-provider";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Strategy Wars: AI-Powered Trading Strategies",
  description: "AI Create - Test - Analyze - Learn: Self-Evolving Trading Strategies",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body suppressHydrationWarning={true}
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
         <Toaster richColors={true} position='top-right'/>
        <NextTopLoader showSpinner={false} color='#61a673' />
        <QueryProvider>
          <WebSocketProvider>
          {children}
          </WebSocketProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
