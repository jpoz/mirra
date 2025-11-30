const providerStyles = {
  gemini: "bg-blue-100 text-blue-800",
  openai: "bg-green-100 text-green-800",
  claude: "bg-orange-100 text-orange-800",
} as const;

export function getProviderStyles(provider: string) {
  if (provider in providerStyles) {
    return providerStyles[provider as keyof typeof providerStyles];
  }
  return "bg-gray-100 text-gray-800";
}
