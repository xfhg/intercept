import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "INTERCEPT",
  description: "Policy as Code Engine",
  lastUpdated: true,
  themeConfig: {
    search: {
      provider: 'local'
    },
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      // { text: 'Home', link: '/' },
      { text: 'Documentation', link: '/docs/tbd' },
      { text: 'Features', link: '/docs/tbd' },
      { text: 'Latest Release', link: 'https://github.com/interceptd/intercept/releases' }
    ],
    // head: [['link', { rel: 'icon', href: '/intercept-icon.svg' }]],
    logo: '/intercept-icon.svg',
    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Architecture', link: '/docs/architecture' },
          { text: 'Platform Build', link: '/docs/platform-build' },
          { text: 'Docker QuickStart', link: '/docs/docker-quickstart' },
          { text: 'Sandbox Playground', link: '/docs/sandbox' },
        ]
      },
      
      {
        text: 'Policy Features',
        items: [
          { text: 'Schema', link: '/docs/policy-schema' },
          { text: 'Enforcement Levels', link: '/docs/enforcement' },
          { text: 'Patching', link: '/docs/patching' },
          { text: 'Monitoring', link: '/docs/monitoring' },
        ]
      },

      {
        text: 'Policy Types',
        items: [
          { text: 'SCAN ', link: '/docs/policy-scan-regex' },
          { text: 'ASSURE ', link: '/docs/policy-assure-regex' },
          { text: 'ASSURE - REGO ', link: '/docs/policy-assure-rego' },
          { text: 'ASSURE - TYPE ', link: '/docs/policy-assure-filetype' },
          { text: 'ASSURE - API ', link: '/docs/policy-assure-api' },
          { text: 'RUNTIME ', link: '/docs/policy-runtime' },
        ]
      },
      {
        text: 'INTERCEPT AUDIT',
        items: [
          { text: 'Compliance Reporting', link: '/docs/compliance-report' },
          { text: 'Feature Flags', link: '/docs/feature-flags' },
        ]
      },
      {
        text: 'INTERCEPT OBSERVE',
        items: [
          { text: 'Setup', link: '/docs/intercept-observe' },
          { text: 'Runtime Daemon', link: '/docs/runtime-observe' },
          { text: 'Integration Webhooks', link: '/docs/hooks' }
        ]
      },
      {
        text: 'Use Cases',
        items: [
          { text: 'Overview', link: '/docs/use-cases' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/xfhg/intercept' }
    ],
    footer: {
      message: 'Released under the <a href="https://github.com/xfhg/intercept/blob/master/LICENSE">EUPL-1.2 License</a>',
      copyright: 'Copyright © 2018-202X - <a href="https://github.com/xfhg">Flávio HG</a>'
    }
  },
  
})
