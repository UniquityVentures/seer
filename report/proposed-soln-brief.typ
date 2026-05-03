#set page(margin: 0.69in)
#set text(size: 12pt)
#set par(justify: true, leading: 0.69em)
#show heading: it => block(above: 1em, below: 0.69em, it)

#align(center)[
  #text(size: 22pt, weight: "bold")[Proposed Solution Brief]
]


= Applicant Details
#table(
    columns: (auto, auto, auto),
    inset: 10pt,
    align: left + horizon,
    "1", [*Applicant Name*], "Raj Tagore",
    "2", [*Startup/MSME Name*], "Uniquity Ventures",
    "3", [*Challenge Title*], [*Problem Statement 18:* AI Based OSINT Analysis and monitoring system and social 
media stacks],
    "4", [*Proposed Duration (in months)*], "12",
    "5", [*Contact and Email ID*], [*Phone*: +91 90043 45953,\ *Email*: uniquityventures\@gmail.com],
)


= Brief Summary of the Proposed Solution
Our proposed solution - Seer is an AI-powered OSINT analysis and monitoring software designed to scrape, collect, filter, and interpret large volumes of global open-source data, while also supporting user-provided custom information domains. It addresses the challenge of analysing fast-growing public data from websites, news sources, social media, images, videos, and other open channels, where manual review is slow, manpower-intensive, and prone to inconsistency.

The architecture of seer is built to run multiple scrapers, source monitors, data processors, and AI analysis jobs in parallel. This makes the platform cost-efficient to operate while still supporting high-throughput collection from websites, social-media platforms, other datasets (GDELT, OpenSky, etc), search-driven DeepSearch workflows, and future custom sources. Its concurrent design is especially suited to OSINT workloads, where many independent sources must be fetched, cleaned, analysed, and refreshed continuously.

The plugin architecture works with a common Intel layer, where extracted material is standardised into searchable intelligence records with AI-generated titles, summaries, source metadata, timestamps, and vector embeddings. New plugins can add specialised information sources, while separate analysis pipelines can process the same information in different ways, such as filtering, summarisation, semantic search, reporting, or GIS-based visualisation. This enables rapid iteration based on user feedback and allows Seer to expand both the breadth of information sources and the depth of analysis without redesigning the core platform.


= Key Technology(s) Used
Golang, Large Language Models, Web Scraping, NLP, Vector Search and RAG, GIS


#pagebreak()


= Deliverables
#table(
    columns: (auto, auto, auto),
    inset: 8pt,
    align: left + horizon,
    [*S.No.*], [*Deliverable Name*], [*Brief Description*],
    "1", [Large scale data scraping engine], [An engine that can scrape large sections of the internet at scale.],
    "2", [OSINT Data Collection Plugins], [A modular plugin system for collecting information from websites, social media platforms, public datasets and any future custom source.],
    "3", [User-Defined Real-Time Monitoring], [Configurable workers that periodically scrape and monitor sources-of-interest based on user-defined parameters.],
    "4", [Data Analysis and Processing Layer], [NLP, AI, and RAG-based processing to clean, classify, summarise, enrich, and analyse gathered data.],
    "5", [Misinformation Identification], [Built-in misinformation identification using AI analysis, cross-source comparison, and user-defined credibility metrics.],
    "6", [Common Intel Datastore], [A unified intelligence database that standardises extracted data from all sources into a common format.],
    "7", [Cross-Source Intelligence Correlation], [Analysis of gathered Intel to discover patterns across different sources helping analysts connect scattered information into a coherent intelligence picture.],
    "8", [Custom Workflows], [Configurable workflows such as deep research, image analysis, object tracking, source-specific filtering, summarisation, similarity search, and report generation.],
    "9", [Managed Intelligence Platform], [A packaged web application with role-based access and tailored UI for intel gatherers, analysts, and senior military decision-makers.]
)
