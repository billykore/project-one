import type { Metadata } from "next";
import { Archivo, Newsreader, IBM_Plex_Mono } from "next/font/google";
import "./globals.css";
import { ErrorModalProvider } from "@/hooks/use-error-modal";
import ErrorModal from "@/components/layout/error-modal";

/* Display and UI: a sturdy grotesque that holds up at small sizes. */
const archivo = Archivo({
  variable: "--font-archivo",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
});

/* Reading: a low-contrast text serif for the words people came for. */
const newsreader = Newsreader({
  variable: "--font-newsreader",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600"],
  style: ["normal", "italic"],
});

/* Margin voice: handles, timestamps, tags, counts, labels. */
const plexMono = IBM_Plex_Mono({
  variable: "--font-plex-mono",
  subsets: ["latin"],
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "Project One",
  description: "Write something. Read what everyone else wrote.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${archivo.variable} ${newsreader.variable} ${plexMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col bg-paper text-ink">
        <ErrorModalProvider>
          {children}
          <ErrorModal />
        </ErrorModalProvider>
      </body>
    </html>
  );
}
