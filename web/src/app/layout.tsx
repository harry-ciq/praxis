import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Providers } from "./providers";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "Praxis — Where execution speaks louder than words",
    template: "%s · Praxis",
  },
  description:
    "Praxis is a professional network where every claim is verified by your work — GitHub commits, YouTube videos, and more. Build a portfolio of real, immutable achievements.",
  applicationName: "Praxis",
  keywords: [
    "portfolio",
    "developer",
    "github",
    "youtube",
    "achievements",
    "verified",
    "professional network",
    "hiring",
  ],
  openGraph: {
    title: "Praxis",
    description:
      "Where execution speaks louder than words. A verified portfolio of your real achievements.",
    siteName: "Praxis",
    type: "website",
    locale: "en_US",
  },
  twitter: {
    card: "summary_large_image",
    title: "Praxis",
    description:
      "Where execution speaks louder than words. A verified portfolio of your real achievements.",
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} dark h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
