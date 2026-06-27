import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { ErrorModalProvider } from "@/hooks/use-error-modal";
import ErrorModal from "@/components/layout/error-modal";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Project One",
  description: "A full-featured social media platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-gray-50 text-gray-900 dark:bg-gray-950 dark:text-gray-100 font-sans">
        <ErrorModalProvider>
          {children}
          <ErrorModal />
        </ErrorModalProvider>
      </body>
    </html>
  );
}
