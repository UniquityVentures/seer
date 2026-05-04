#import "@preview/fletcher:0.5.8" as fletcher: diagram, node, edge
#set heading(numbering: "1.")

#set page(
  margin: (
    top: 6.9em,
    bottom: 4.2em,
    left: 4.2em,
    right: 4.2em,
  ),
  header: [
    #align(right)[
      #block([
        #image("logo.webp")
      ], below: -2em, height: 4em)
    ],
  ],
  footer: context align(center)[#counter(page).display("1")]
)

#set text(size: 12pt)
#set par(justify: true, leading: 0.65em)
#show heading: it => block(above: 1em, below: 0.69em, it)


#align(center)[
  #text(size: 22pt, weight: "bold")[Proposed Technical Solution]
]

*For Problem Statement 18:* AI Based OSINT Analysis and monitoring system and social media stacks

#pagebreak()

#outline()

#pagebreak()


= Framing the Problem Statement

This section explains why OSINT frameworks are difficult to build and maintain, and uses those challenges to define the minimum requirements of a serious solution. A good OSINT platform must not only collect public data, but also survive unstable sources, extract source-specific context, filter irrelevant information early, process multi-modal evidence, and remain adaptable as the web and mission requirements change.


== Web Sources are Adversarial and Volatile

Modern OSINT systems depend on public platforms such as social networks, forums, media sites, and open web sources. These platforms are not stable data providers. Their layouts, identifiers, request patterns, and access rules change frequently, and many of them actively resist automated collection.

This creates an adversarial scraping environment. Scrapers may fail because of DOM changes, rate limits, bot-detection systems, browser fingerprinting, behavioural analysis, IP reputation checks, CAPTCHAs, login restrictions, or changes in how content is rendered. A collection method that works today may stop working in a few weeks. Any serious OSINT framework must therefore assume that source access will degrade, fail, and need continuous maintenance.

This is why scraper logic must be decoupled from the core platform. If every source fetcher is tightly embedded inside the main application, the entire software becomes expensive to update whenever one website changes. A better design separates fetchers into independently maintainable services, so individual source collectors can be repaired, replaced, scaled, or redeployed without disturbing the rest of the system.

== Source Data is Not Uniform

OSINT value is rarely contained only in the visible text of a page. Each platform stores useful context differently. A post may include comments, media metadata, author information, timestamps, location signals, engagement patterns, linked pages, or embedded content. What matters for one source may be irrelevant for another.

This makes generic scraping insufficient. A Reddit post, an Instagram post, a Facebook thread, a LinkedIn profile, and a news article each require different extraction logic. The system must understand the source type, preserve the source-native record, and then extract intelligence-ready content using source-specific processing rules.

== Useful Intelligence is Buried in Noise

The open web contains far more irrelevant information than useful intelligence. A broad collection system will quickly encounter entertainment content, memes, advertisements, duplicate posts, low-value commentary, and unrelated public discussion. If this material is stored without filtering, the database becomes noisy and downstream analysis becomes less reliable.

For this reason, filtering must happen early in the pipeline. The system should remove irrelevant data before it populates the Intel layer, while preserving enough raw evidence for traceability. Clean source ingestion is essential for reliable search, alerting, reporting, geospatial analysis, and AI-assisted interpretation.

== OSINT is Multi-Modal and Contextual

Online intelligence is not limited to text. A single record may contain images, videos, captions, comments, links, locations, timestamps, and metadata. These elements may only become meaningful when analysed together. For example, the caption of a post, the location attached to it, the comments below it, and the media metadata may each contribute part of the intelligence picture.

An OSINT framework must therefore support multi-modal processing, source-aware metadata extraction, geotagging, and cross-source correlation. Without this, valuable signals remain trapped inside media files, comments, links, or platform-specific fields that a generic text scraper would miss.

== Maintainability is Key

Most OSINT implementations struggle because they are built as static tools for a dynamic environment. The challenge is not only to build scrapers, models, and dashboards once, but to keep them working as the web changes. The system must be designed for continuous maintenance from the beginning.

This requires modular architecture, loose coupling, component isolation, defined interfaces, independent deployment, and least-privilege development. A fetcher microservice can be maintained by a team that only has access to that source collector, without exposing the main software, analyst workflows, or sensitive operational data. This makes the platform safer to maintain and more realistic to operate over time.


#figure(
  align(center)[
    #set par(justify: false)
    #text(font: ("DejaVu Sans"), size: 9pt)[
  #diagram(
    node-stroke: 0.8pt,
    edge-stroke: 0.9pt,
    spacing: (1.3em, 1.1em),
    node((0, 0), align(center)[
      #text(weight: "bold")[OSINT Stage]
    ], width: 4.2cm, inset: 7pt, corner-radius: 4pt, fill: blue.lighten(88%)),
    node((1, 0), align(center)[
      #text(weight: "bold")[Challenge Faced]
    ], width: 8.2cm, inset: 7pt, corner-radius: 4pt, fill: red.lighten(88%)),

    node((0, 1), align(center)[Public web collection], width: 4.2cm, inset: 7pt, corner-radius: 4pt),
    edge((0, 1), (1, 1), "->"),
    node((1, 1), align(center)[
      Layout changes, bot detection, rate limits, browser fingerprinting, and CAPTCHAs can break scrapers quickly.
    ], width: 8.2cm, inset: 7pt, corner-radius: 4pt, fill: red.lighten(94%)),

    node((0, 2), align(center)[Source extraction], width: 4.2cm, inset: 7pt, corner-radius: 4pt),
    edge((0, 2), (1, 2), "->"),
    node((1, 2), align(center)[
      Each platform hides useful context differently, including comments, metadata, links, location, and media.
    ], width: 8.2cm, inset: 7pt, corner-radius: 4pt, fill: orange.lighten(90%)),

    node((0, 3), align(center)[Relevance filtering], width: 4.2cm, inset: 7pt, corner-radius: 4pt),
    edge((0, 3), (1, 3), "->"),
    node((1, 3), align(center)[
      Valuable intelligence is buried inside large volumes of irrelevant, duplicate, or low-quality public data.
    ], width: 8.2cm, inset: 7pt, corner-radius: 4pt, fill: yellow.lighten(86%)),

    node((0, 4), align(center)[Multimodal analysis], width: 4.2cm, inset: 7pt, corner-radius: 4pt),
    edge((0, 4), (1, 4), "->"),
    node((1, 4), align(center)[
      Text, images, video, captions, metadata, and geotags must be interpreted together to produce useful Intel.
    ], width: 8.2cm, inset: 7pt, corner-radius: 4pt, fill: green.lighten(88%)),

    node((0, 5), align(center)[Long-term operation], width: 4.2cm, inset: 7pt, corner-radius: 4pt),
    edge((0, 5), (1, 5), "->"),
    node((1, 5), align(center)[
      The system must remain maintainable while sources, missions, and access constraints continue to change.
    ], width: 8.2cm, inset: 7pt, corner-radius: 4pt, fill: purple.lighten(90%)),
  )
  ]
  ],
  caption: [Key challenges faced at each stage of an OSINT framework],
)

The central problem, therefore, is not merely data collection. It is building an OSINT framework that can survive source instability, extract source-specific context, filter noise early, process multi-modal evidence, and remain maintainable across the constantly changing landscape of the web.

#pagebreak()


= Technical Architecture and Approach
Seer is proposed as a distributed, large scale OSINT analysis and monitoring platform. The system is designed to collect large volumes of public and user-provided information through source plugins, store it in a common Intel layer, process it through AI-assisted analysis plugins, and present actionable outputs through a managed role-based application. It directly addresses the problem statement by reducing manual collection effort, enabling real-time monitoring of user-defined subjects, and improving the reliability, traceability, and usability of intelligence outputs.

As we saw above, OSINT frameworks are difficult to build because the problem does not stay fixed. Sources change formats, websites break scrapers, platforms impose rate limits, data arrives with noise and misinformation, and analyst requirements shift with each mission. Seer is designed for this instability. 
Seer's plugin-based architecture provides modularity, separation of concerns, and component isolation. Parts of the software can be upgraded independently without modifying the rest of the platform.
This adaptability gives Seer a stronger chance of success than a fixed OSINT dashboard because it can scale and evolve as sources, threats, and operational needs change.

The proposed system is not starting from a purely theoretical design. A working prototype of Seer already exists and demonstrates the core architecture: a Reddit source, website scraper, deep research AI agent, AI chat interface, Intel layer, multimodal embedding support, Retrieval Augmented Generation (RAG), scheduled scraping workers, and a GIS map. The grant would therefore be used to harden and scale an existing architecture rather than to begin exploratory development from zero.

The current prototype validates the core direction of the platform. The proposed work will focus on preparing it for larger and more demanding use: scaling collection across more sources, improving source-specific plugins, strengthening the Intel layer for military requirements, adding stronger reliability and misinformation checks, and preparing for deployment needs such as rate limits, unstable sources, distributed workers, access control, and audit logs.


== High Level Overview

Seer follows a layered, plugin-based architecture. A fleet of data-scraping microservices collects public web data. Other services, including public datasets, API-accessible sources, and proprietary services, can feed the platform through the same source-plugin interface. 

Each source is represented by an independent plugin, so new source types can be added without changing the rest of the system.
Within each source plugin, collected material passes through a common pipeline for data fetching, pre-processing, filtering, raw data storage, and extraction of content suitable for intelligence generation. Source-specific stores preserve the original or source-native records.

The Intel layer processes the extracted content into a defined format and feeds it into a shared model before saving it to the database. Downstream applications then consume this Intel model for semantic search, AI chat, similarity search, report generation, interactive maps, and alerts and warnings, while retaining links back to the underlying source records where available.

#figure(
  image("Seer Architecture.svg", height: 72%),
  caption: [Seer high-level system architecture],
)

== Data Scraping Microservice Fleet

The first layer of Seer is a fleet of independent data-scraping microservices. Instead of placing all scraping logic inside the main application, each major web source can have its own dedicated fetcher, such as a Reddit scraper, website scraper, Twitter scraper, or future source-specific scraper. This keeps the data collection layer operationally separate from the core Seer application. If one fetcher is rate-limited, blocked, banned, or broken because a website changes its layout, the rest of the platform continues to function.

This microservice architecture also allows Seer to implement redundancy at the correct level. Public web sources frequently use bot-detection systems, rate limits, suspicious traffic detection, fingerprinting, CAPTCHA checks, and behavioural analysis. Seer can respond to these constraints using proxy rotation, browser spoofing, fingerprint randomisation, CAPTCHA solving, and human-behaviour emulation. These capabilities can be shared across the fleet while still allowing source-specific handling. For example, the safeguards needed for a social media platform may be different from the safeguards needed for a news website, forum, or public search page.

Separating scrapers into independent services also improves scale and maintainability. Large-scale fetching can run in parallel across many workers without overloading the main application. Scrapers can be deployed, restarted, scaled, or replaced independently. Different developer teams can maintain different fetchers without needing full access to the main application, unrelated source plugins, analyst workflows, or sensitive operational data. This reduces security and privacy risk while allowing the collection layer to evolve quickly as external websites and platforms change.

#figure(
  align(center)[
    #set par(justify: false)
    #text(font: ("DejaVu Sans"), size: 9pt)[
  #diagram(
    node-stroke: 0.8pt,
    edge-stroke: 0.9pt,
    spacing: (1.2em, 1em),
    node((0, 0), align(center)[Reddit\ Fetcher], width: 3cm, inset: 6pt, corner-radius: 4pt, fill: orange.lighten(88%)),
    node((0, 1), align(center)[Website\ Fetcher], width: 3cm, inset: 6pt, corner-radius: 4pt, fill: orange.lighten(88%)),
    node((0, 2), align(center)[Social Media\ Fetcher], width: 3cm, inset: 6pt, corner-radius: 4pt, fill: orange.lighten(88%)),

    node((1, 1), align(center)[
      #text(weight: "bold")[Source Plugin Interface]\
      #text(size: 8pt)[defined contract for fetched data]
    ], inset: 7pt, corner-radius: 4pt, fill: blue.lighten(88%)),

    node((2, 1), align(center)[
      #text(weight: "bold")[Core Seer Platform]\
      #text(size: 8pt)[Intel layer, analysis, workflows]
    ], width: 4.2cm, inset: 7pt, corner-radius: 4pt, fill: green.lighten(88%)),

    edge((0, 0), (1, 1), "->"),
    edge((0, 1), (1, 1), "->"),
    edge((0, 2), (1, 1), "->"),
    edge((1, 1), (2, 1), "->"),

    node((0, 3), align(center)[
      #text(size: 8pt)[independent deployment \ isolated maintenance \ least-privilege access]
    ], inset: 6pt, corner-radius: 4pt, fill: gray.lighten(90%)),
  )
  ]
  ],
  caption: [Fetcher microservices remain decoupled from the core Seer platform],
)

== Other Services

The architecture is not limited to web scraping. Seer can also ingest information from other programmatic services, including public datasets, API-accessible sources, proprietary services, and future private data feeds. These services enter the same source-plugin workflow as scraper-backed sources, which means the platform can treat a structured API, an open government dataset, an aviation feed, a maritime feed, or an internal proprietary source as another source of intelligence.

This design gives Seer a single ingestion model for many kinds of data. A source may be fetched by a scraper microservice, queried through an authenticated API, received from a public dataset, or connected through a proprietary integration, but once it reaches the source-plugin layer it can be processed, filtered, stored, and converted into Intel using the same architectural pattern. This makes it possible to add new sources without redesigning downstream search, reporting, maps, AI chat, or alerting features. It also allows Seer to run multiple data fetchers for a single source type, ensuring coverage for sources that require strict full-time monitoring even if one fetcher goes down, is rate-limited, or temporarily loses access.

== Source Plugins

Source Plugins form the main integration layer between external information and Seer's Intel layer. Each plugin represents a specific type of source, such as Reddit, websites, Twitter, OpenSky, GDELT, BlueSky, or a future custom source. The plugin is responsible for knowing how that source should be fetched, pre-processed, filtered, stored, and converted into intelligence-ready content.

=== Data Fetching and Cross-Plugin Processing

The first stage inside a Source Plugin is data fetching. A plugin may call one of the scraping microservices, query an external API, pull from a public dataset, or connect to a proprietary service. Plugins can also define workers that run on specified schedules. For example, a Reddit source may refresh selected subreddits every 25 minutes, a website source may crawl a list of news sites every hour, and a public dataset source may poll for new records on a fixed cadence.

Seer also supports cross-plugin execution through a proprietary capability called *SourceWeave*. SourceWeave allows one plugin to invoke selected functions from another plugin when a source contains embedded or linked content. For example, if the Reddit Source Plugin finds a URL inside a post or comment, it can call the Website Source Plugin's fetcher to parse the linked webpage and feed that extracted content into the same intelligence pipeline. This enables real multi-source and multi-modal processing without forcing every plugin to reimplement every other plugin's logic.

=== Pre-Processing and Filtering

After fetching, each plugin runs source-specific pre-processing. This is important because different sources do not produce data in the same shape. A Reddit post requires different handling from a website article, a public event record, an aviation feed, or a social media thread. Reddit may require post parsing, comment traversal, subreddit metadata, author identity, and vote or engagement signals. A website may require HTML cleanup, boilerplate removal, article extraction, title detection, canonical URL handling, and publication-date parsing. By separating Source Plugins from the rest of the system, Seer can handle these differences natively.

Pre-processing also includes deduplication. When Seer monitors the same source repeatedly, it will naturally encounter data it has already fetched. The plugin can compare incoming records against its source-specific raw data store and keep only unique or updated material. This prevents repeated posts, unchanged pages, duplicate events, and already-processed records from polluting the Intel layer.

The filtering stage allows users to specify what information they are looking for inside a source. These filters can be written in natural language or implemented as scripts. Natural-language filters are useful for analyst workflows because users can describe the subject of interest directly, such as "military movement near a border region" or "reports related to a specific conflict". Seer can then apply LLM-based filtering logic to discard irrelevant material before it becomes Intel.

Filtering also mitigates the weakness of generic search-engine results and broad source feeds. A news website may publish important defence reporting alongside entertainment news, celebrity articles, advertisements, sports updates, and unrelated opinion pieces. Relevant information can be buried among this noise. The Source Plugin's filtering logic helps keep only material that matches the user's objective. When users need more granular control, they can write custom filtering scripts to enforce precise rules.

=== Raw Storage and Intel Extraction

The raw data storage stage preserves source-native records in source-specific databases, such as RedditKind, WebsiteKind, or GDELTKind. This storage provides an audit log of what data was fetched, when it was fetched, where it came from, and how it was represented before Intel extraction. It also strengthens deduplication because each plugin can compare new records against its own historical store.

Raw data storage is also important for traceability and misinformation handling. If Seer later identifies incorrect Intel, suspicious claims, or possible misinformation, analysts can trace that Intel back to the exact source record that produced it. This helps determine whether the issue came from a bad extraction, an unreliable source, a compromised feed, or deliberate misinformation. Over time, this source-level traceability can help identify compromised or low-quality sources earlier.

Finally, each plugin extracts content for Intel generation. The plugin reads the source-native record and produces the information needed by the Intel layer, including content, metadata, datetimes, poster or publisher identity, links, media references, and source-specific context. This logic remains source-specific because the useful metadata of a Reddit post is different from the useful metadata of a blog post, a LinkedIn profile, or a public event record. Plugins can also include custom extraction logic such as reading Reddit comments, following links, inspecting LinkedIn connections, parsing media metadata, or extracting structured fields from public datasets.

The result is a flexible source layer that can fetch data, remove noise, retain evidence, and produce intelligence-ready content while preserving the ability to evolve each source independently.

#figure(
  align(center)[
    #set par(justify: false)
    #text(font: ("DejaVu Sans"), size: 9pt)[
  #diagram(
    node-stroke: 0.8pt,
    edge-stroke: 0.9pt,
    spacing: (1em, 1em),

    node((0, 0), align(center)[
      #text(weight: "bold")[Sources Fetch]\
      #text(size: 8pt)[Scrapers \ APIs \ Datasets \ Feeds]
    ], width: 4cm, inset: 6pt, corner-radius: 4pt, fill: gray.lighten(90%)),

    node((1, 0), align(center)[Pre-process\ + Dedup], width: 2.6cm, inset: 6pt, corner-radius: 4pt, fill: yellow.lighten(85%)),
    node((2, 0), align(center)[Filter\ (NL / scripts)], width: 2.6cm, inset: 6pt, corner-radius: 4pt, fill: green.lighten(88%)),
    node((3, 0), align(center)[
      #text(weight: "bold")[Raw Store]\
      #text(size: 8pt)[source-native record\ audit history \ dedup history]
    ], width: 3.2cm, inset: 6pt, corner-radius: 4pt, fill: red.lighten(92%)),
    edge((1, 0), (1, 0), "->"),
    edge((1, 0), (0, 0), "..>", bend: 90deg, label: text(size: 7.5pt)[Source Weave]),
    node((4, 0), align(center)[Intel\ Extraction], width: 2.6cm, inset: 6pt, corner-radius: 4pt, fill: blue.lighten(88%)),

    edge((0, 0), (1, 0), "->"),
    edge((1, 0), (2, 0), "->"),
    edge((2, 0), (3, 0), "->"),
    edge((3, 0), (4, 0), "->"),

  )
  ]
  ],
  caption: [Inside a Source Plugin: fetch, pre-process, filter, store, and extract Intel, with SourceWeave for cross-plugin calls],
)

== Intel Layer

The Intel Layer is responsible for converting the source-specific material produced by Source Plugins into a common intelligence model. Source Plugins may produce raw posts, webpages, comments, event records, API responses, media files, and other source-native objects. The Intel Layer extracts the most vital information from this material and reformats it into a standard Intel record that downstream applications can search, compare, analyse, and visualise.

This extraction process uses LLMs to identify and structure the important parts of the raw content. Depending on the source, the Intel Layer can extract a title, description, datetime, source kind, relevant entities, locations, events, summaries, claims, and other metadata needed for analysis. The goal is not merely to store raw content, but to convert it into a consistent intelligence object that can be used reliably across the platform.

Every Intel record preserves a link back to the original raw data stored by the Source Plugin. This creates a clear origin trail for each piece of intelligence: what source produced it, when it was fetched, what raw record it came from, and how it entered the Intel database. This traceability is essential for audit, verification, and analyst confidence.

The same origin trail also supports misinformation and source-integrity analysis. If two sources feed Seer conflicting Intel, analysts can trace both claims back to their raw records and source histories. Over time, this helps identify unreliable, compromised, or deliberately misleading sources. Instead of treating Intel as isolated text, Seer can evaluate where it came from and whether the source has a pattern of producing conflicting or low-confidence information.

After the Intel record is structured, Seer generates vector embeddings for the content. For multimodal material, including text, images, and other media where available, Seer can use an open-source multimodal embedding model to represent the content in vector space. These embeddings power downstream applications such as semantic search, similarity search, retrieval for AI chat, clustering, correlation, and map or report workflows.

#figure(
  align(center)[
    #set par(justify: false)
    #text(font: ("DejaVu Sans"), size: 9pt)[
  #diagram(
    node-stroke: 0.8pt,
    edge-stroke: 0.9pt,
    spacing: (1.1em, 1em),

    node((0, 0), align(center)[
      #text(weight: "bold")[Source Plugins]\
      #text(size: 8pt)[raw posts, pages, feeds, media]
    ], width: 3.6cm, inset: 6pt, corner-radius: 4pt, fill: orange.lighten(88%)),

    node((1, 0), align(center)[
      #text(weight: "bold")[LLM Extraction]\
      #text(size: 8pt)[title, summary, entities,\ locations, events, claims]
    ], width: 4cm, inset: 6pt, corner-radius: 4pt, fill: yellow.lighten(85%)),

    node((2, 0), align(center)[
      #text(weight: "bold")[Intel Record]\
      #text(size: 8pt)[shared model \ provenance link]
    ], width: 4cm, inset: 6pt, corner-radius: 4pt, fill: blue.lighten(88%)),

    node((3, 0), align(center)[
      #text(weight: "bold")[Multimodal Embedding]\
      #text(size: 8pt)[text + image + media\ in shared vector space]
    ], width: 4cm, inset: 6pt, corner-radius: 4pt, fill: green.lighten(88%)),

    edge((0, 0), (1, 0), "->"),
    edge((1, 0), (2, 0), "->"),
    edge((2, 0), (3, 0), "->"),

    node((0, 1), align(center)[
      #text(weight: "bold")[Raw Store]\
      #text(size: 8pt)[source-native record]
    ], width: 3.2cm, inset: 6pt, corner-radius: 4pt, fill: red.lighten(92%)),
    edge((0, 0), (0, 1), "->"),
    edge((2, 0), (0, 1), "..>", bend: 25deg, label: text(size: 7.5pt)[provenance trail]),

    // node((3, 1), align(center)[
    //   #text(weight: "bold")[Downstream Applications]\
    //   #text(size: 8pt)[search | chat | reports | maps | alerts]
    // ], inset: 6pt, corner-radius: 4pt, fill: purple.lighten(88%)),
    // edge((3, 0), (3, 1), "->"),
  )
  ]
  ],
  caption: [The Intel Layer normalises source records into a shared, traceable, embedding-backed intelligence model],
)

== Downstream Applications

The architecture above enables several downstream applications to operate on the same Intel model. Because content from different sources is normalised into a shared structure and embedded into a common vector space, Seer can search and analyse information across websites, social platforms, public datasets, APIs, and proprietary feeds at the same time.

#figure(
  align(center)[
    #set par(justify: false)
    #text(font: ("DejaVu Sans"), size: 9pt)[
      #diagram(
        node-stroke: 0.8pt,
        edge-stroke: 0.9pt,
        spacing: (1em, 1.15em),
        node((1, 1), align(center)[
          #text(weight: "bold")[Intel Layer]\
          #text(size: 8pt)[shared Intel model + embeddings]
        ], width: 3.4cm, inset: 7pt, corner-radius: 4pt, fill: blue.lighten(85%)),
        node((1, 0), align(center)[
          Semantic \&\ similarity search
        ], width: 3cm, inset: 6pt, corner-radius: 4pt, fill: yellow.lighten(88%)),
        node((0, 1), align(center)[
          Cross-source correlation
        ], width: 2.8cm, inset: 6pt, corner-radius: 4pt, fill: green.lighten(90%)),
        node((2, 1), align(center)[
          AI analyst chat
        ], width: 2.8cm, inset: 6pt, corner-radius: 4pt, fill: orange.lighten(90%)),
        node((0, 2), align(center)[
          Evidence-linked reports
        ], width: 2.8cm, inset: 6pt, corner-radius: 4pt, fill: purple.lighten(92%)),
        node((1, 2), align(center)[
          Geospatial maps
        ], width: 2.8cm, inset: 6pt, corner-radius: 4pt, fill: rgb("#c8e6c9")),
        node((2, 2), align(center)[
          Alerts \& warnings
        ], width: 2.8cm, inset: 6pt, corner-radius: 4pt, fill: red.lighten(92%)),
        edge((1, 1), (1, 0), "->"),
        edge((1, 1), (0, 1), "->"),
        edge((1, 1), (2, 1), "->"),
        edge((1, 1), (0, 2), "->"),
        edge((1, 1), (1, 2), "->"),
        edge((1, 1), (2, 2), "->"),
      )
    ]
  ],
  caption: [The shared Intel model powers multiple downstream applications from one consistent data and embedding layer],
)

=== Semantic and Similarity Search

Semantic search allows analysts to search the Intel database by meaning rather than only by exact keywords. Similarity search uses Intel embeddings to find records that are semantically or visually close to a selected item. This helps surface related posts, articles, events, images, or reports even when they use different wording or come from different source types.

=== Cross-Source Correlation

The shared Intel structure also allows Seer to identify connections across sources and across time. For example, a social media post, a news article, a public event record, and an aviation or maritime signal may each describe part of the same activity. By analysing them together, Seer can help reveal patterns that would be difficult to detect inside a single feed. This is especially valuable for monitoring evolving situations where weak signals appear gradually across multiple sources.

=== AI Analyst Chat

Seer also includes AI tools such as an analyst chatbot that can query the Intel database and answer questions using the collected evidence. The chatbot can be equipped to trigger new source fetches, inspect gaps in the available intelligence, and add newly discovered sources into the monitoring workflow. This turns the assistant from a passive question-answering tool into an active interface for expanding collection.

=== Evidence-Linked Report Generation

A report generator has also been implemented. It can query the Intel gathered about a requested topic and produce a detailed, comprehensive report suitable for review by officials. Because the report is generated from structured Intel records linked to raw source material, the output can retain evidence trails rather than becoming an untraceable narrative summary.

=== Alerts and Warnings

Seer can also support alert and warning systems. Analysts can define alert conditions in natural language, such as the appearance of a specific threat pattern, location, source combination, or operational signal. Whenever newly added Intel matches those conditions, the alert can trigger automatically. For users who require more control, the same alerting system can support scripted conditions.

=== Geospatial Visualisation

Geospatial visualisation is another downstream application. Intel with location metadata can be plotted onto an interactive map to help analysts identify spatial patterns, clusters, movement, proximity, and regional changes over time. This can support situational awareness, field planning, and the discovery of geographic relationships between otherwise disconnected pieces of Intel.

=== Mission-Specific Intelligence Applications

The same architecture can support additional applications such as person-of-interest tracking, knowledge graphs, source reliability dashboards, timeline reconstruction, entity correlation, and automated briefings. The important point is that Seer's architecture does not lock the system into a single dashboard or workflow. Once the military's requirements, operational procedures, and existing systems are studied in detail, the same source-plugin and Intel-layer foundation can be used to build the downstream tools that best fit the mission.

#pagebreak()

= Innovation

Seer's innovation lies in how it combines resilient source collection, source-aware AI processing, provenance-based intelligence analysis, and agentic downstream workflows into one extensible OSINT architecture. The following capabilities distinguish it from a basic OSINT dashboard or a conventional scraping-and-search system.

== SourceWeave: Multi-modal and multi-source data processing

SourceWeave is Seer's cross-plugin processing mechanism. It allows one Source Plugin to invoke selected functions from another Source Plugin when the incoming data contains linked, embedded, or dependent material. For example, if the Reddit Source Plugin encounters a URL inside a post or comment, it can call the Website Source Plugin to fetch, parse, and process that page before feeding both pieces of evidence into the Intel Layer.

This is more powerful than treating each source as an isolated feed. Real OSINT data is often nested across platforms: a post links to a website, a website references a social account, a profile links to media, and a dataset points to external records. SourceWeave allows Seer to follow these relationships through controlled plugin-to-plugin calls, enabling multi-source and multi-modal intelligence extraction without duplicating parsing logic across plugins.

#figure(
  align(center)[
    #set par(justify: false)
    #text(font: ("DejaVu Sans"), size: 9pt)[
      #diagram(
        node-stroke: 0.8pt,
        edge-stroke: 0.9pt,
        spacing: (1.2cm, 1cm),

        node((1, 0), align(center)[
          #text(weight: "bold")[Incoming Reddit Post]\
          #text(size: 7.5pt)[text + linked URL + embedded image]
        ], width: 5.5cm, inset: 6pt, corner-radius: 4pt, fill: red.lighten(90%)),

        node((1, 1), align(center)[
          #text(weight: "bold")[Reddit Plugin]\
          #text(size: 7.5pt)[entry plugin]
        ], width: 3cm, inset: 6pt, corner-radius: 4pt, fill: orange.lighten(88%)),

        node((1, 2), align(center)[
          #text(weight: "bold")[SourceWeave Mesh]\
          #text(size: 7.5pt)[any plugin can invoke any other]
        ], width: 4cm, inset: 7pt, corner-radius: 8pt, fill: gray.lighten(82%)),

        node((0, 2), align(center)[
          #text(weight: "bold")[Website]\
          #text(size: 7.5pt)[parses URL]
        ], width: 2.4cm, inset: 6pt, corner-radius: 4pt, fill: green.lighten(88%)),

        node((2, 2), align(center)[
          #text(weight: "bold")[Media]\
          #text(size: 7.5pt)[reads metadata]
        ], width: 2.4cm, inset: 6pt, corner-radius: 4pt, fill: purple.lighten(90%)),

        node((0, 3), align(center)[
          #text(weight: "bold")[Twitter]
        ], width: 2.4cm, inset: 6pt, corner-radius: 4pt, fill: yellow.lighten(85%)),

        node((2, 3), align(center)[
          #text(weight: "bold")[Dataset]
        ], width: 2.4cm, inset: 6pt, corner-radius: 4pt, fill: blue.lighten(92%)),

        node((1, 4), align(center)[
          #text(weight: "bold")[Intel Layer]\
          #text(size: 7.5pt)[combined multi-source, multi-modal Intel]
        ], width: 5.5cm, inset: 7pt, corner-radius: 4pt, fill: blue.lighten(85%)),

        edge((1, 0), (1, 1), "->"),
        edge((1, 1), (1, 2), "<->", label: text(size: 7pt)[invokes]),
        edge((1, 2), (0, 2), "<->"),
        edge((1, 2), (2, 2), "<->"),
        edge((1, 2), (0, 3), "<->", bend: 15deg),
        edge((1, 2), (2, 3), "<->", bend: -15deg),
        edge((1, 2), (1, 4), "->", label: text(size: 7pt)[combined evidence]),
      )
    ]
  ],
  caption: [SourceWeave forms a mesh between Source Plugins so a single entry plugin can use other plugins to fully process nested, multi-source content],
)

== Adaptive Fetcher Camouflage

Different platforms use different anti-bot and abuse-detection systems. Seer therefore treats anti-blocking logic as a source-specific capability rather than a generic scraper setting. Fetchers can use proxy rotation, browser spoofing, fingerprint randomisation, CAPTCHA handling, and human-behaviour emulation in ways tailored to the source being monitored.

This allows Seer to adapt to the defensive behaviour of each target platform. A news website, social media platform, forum, public search page, or structured dataset may each require a different collection strategy. By isolating these strategies inside fetcher microservices, Seer can harden collection without exposing or destabilising the rest of the system.

== Selective data fetching

Seer allows users to define what they are looking for at the source level, before irrelevant material becomes Intel. Analysts can describe the desired information in natural language, while advanced users can provide scripts for stricter filtering requirements. This hybrid filtering model helps remove advertisements, unrelated articles, broad search noise, and irrelevant social media content before it enters the Intel database.

The innovation is that filtering happens as part of the source pipeline, not only after data has already been collected and indexed. This reduces noise in downstream search, reports, alerts, and AI responses, while still giving technical users the ability to enforce precise rules when required.

== Source credibility tracking

Every Intel record remains linked to the raw source record that produced it. Beyond auditability, this enables Seer to compare conflicting claims by tracing them back to their original sources. If multiple sources repeatedly produce contradictory or low-confidence Intel, Seer can help analysts identify unreliable feeds, compromised sources, or coordinated misinformation patterns.

This turns provenance from simple record-keeping into an analytical capability. The system can support source credibility assessment by examining not only what was claimed, but where the claim originated and whether that source has a history of conflict with other evidence.

#figure(
  align(center)[
    #set par(justify: false)
    #text(font: ("DejaVu Sans"), size: 9pt)[
      #diagram(
        node-stroke: 0.8pt,
        edge-stroke: 0.9pt,
        spacing: (1em, 1em),

        node((0, 0), align(center)[
          #text(weight: "bold")[Source A]\
          #text(size: 7.5pt)[history: reliable]
        ], width: 2.8cm, inset: 6pt, corner-radius: 4pt, fill: green.lighten(90%)),
        node((1, 0), align(center)[
          #text(weight: "bold")[Source B]\
          #text(size: 7.5pt)[history: reliable]
        ], width: 2.8cm, inset: 6pt, corner-radius: 4pt, fill: green.lighten(90%)),
        node((2, 0), align(center)[
          #text(weight: "bold")[Source C]\
          #text(size: 7.5pt)[history: prior conflicts]
        ], width: 2.8cm, inset: 6pt, corner-radius: 4pt, fill: red.lighten(92%)),

        node((0, 1), align(center)[#text(size: 8pt)[Claim: "X happened"]], width: 2.8cm, inset: 5pt, corner-radius: 4pt),
        node((1, 1), align(center)[#text(size: 8pt)[Claim: "X happened"]], width: 2.8cm, inset: 5pt, corner-radius: 4pt),
        node((2, 1), align(center)[#text(size: 8pt)[Claim: "X did not happen"]], width: 2.8cm, inset: 5pt, corner-radius: 4pt),

        edge((0, 0), (0, 1), "->"),
        edge((1, 0), (1, 1), "->"),
        edge((2, 0), (2, 1), "->"),

        node((1, 2), align(center)[
          #text(weight: "bold")[Intel Layer]\
          #text(size: 7.5pt)[provenance check + conflict detection]
        ], width: 6cm, inset: 7pt, corner-radius: 4pt, fill: blue.lighten(86%)),

        edge((0, 1), (1, 2), "->"),
        edge((1, 1), (1, 2), "->"),
        edge((2, 1), (1, 2), "->"),

        node((0, 3), align(center)[
          #text(weight: "bold")[Verified Intel]\
          #text(size: 7.5pt)[A + B agree, history clean]
        ], width: 3.4cm, inset: 6pt, corner-radius: 4pt, fill: green.lighten(85%)),
        node((2, 3), align(center)[
          #text(weight: "bold")[Quarantined Claim]\
          #text(size: 7.5pt)[C flagged: low credibility]
        ], width: 3.4cm, inset: 6pt, corner-radius: 4pt, fill: red.lighten(85%)),

        edge((1, 2), (0, 3), "->"),
        edge((1, 2), (2, 3), "->"),

        node((1, 4), align(center)[
          #text(weight: "bold")[Analyst]\
          #text(size: 7.5pt)[receives verified Intel + misinformation warning]
        ], width: 6cm, inset: 7pt, corner-radius: 4pt, fill: yellow.lighten(85%)),

        edge((0, 3), (1, 4), "->"),
        edge((2, 3), (1, 4), "..>", label: text(size: 7pt)[warning]),
      )
    ]
  ],
  caption: [Provenance-backed conflict detection acts as a preemptive shield against misinformation reaching the analyst],
)

== Multimodal Intel Embedding

Seer can represent text, images, and other supported media in a shared vector space using open-source multimodal embedding models. This allows downstream applications to search and compare intelligence by semantic and visual similarity, not just by keywords or source type.

For OSINT workflows, this is useful when related evidence appears in different forms: an image, a short post, a long article, and a structured dataset may all refer to the same event. Multimodal embeddings allow these signals to be discovered together.

#pagebreak()

= Implementation and Feasibility

== Working Prototype
The proposed implementation is feasible because Seer is not a concept-only proposal. A *Working Prototype* has been tested as a proof-of-concept and demonstrates the core platform capabilities needed for this solution: source plugins, Reddit and website collection, scheduled scraping workers, an Intel layer, multimodal embedding support, RAG-based AI chat, deep research workflows, report generation, and GIS visualisation. 


#text(style: "italic", weight: "bold")[You can preview a prototype of the proposed solution - Seer at https://seer.lariv.in/.]

#figure(
  image("Seer HomePage.png"),
  caption: [A screenshot of a demo of the proposed software - Seer, in action],
)

== What a full-scale deployment would require

The current prototype is a proof of concept that demonstrates the core architecture end to end, but it runs on a small number of nodes, calls external model APIs for chat and embeddings, and relies on the developers themselves for ad-hoc fetcher maintenance. A full-scale deployment closes that gap. The three requirements that drive the budget below are:

+ *Distributed fetcher fleet across regions and networks:* Public sources actively detect and block automated collection through IP reputation, fingerprinting, rate limits, and behavioural signals. A production deployment needs several fetcher nodes spread across different regions, networks, and hosting providers, so that a block on one node or one network does not interrupt collection and so each source plugin can be routed through the network conditions that work for it.

+ *A team of engineers to maintain and upgrade the fetcher fleet:* Source platforms change layouts, identifiers, anti-bot defences, and access rules continuously. A fetcher that works today can break in a week, and new sources are added as analyst priorities shift. The platform therefore needs a steady team responsible for monitoring fetcher health, repairing scrapers when sources change, adding new source plugins, and maintaining the operational tooling around the fleet.

+ *On-premise multi-GPU server for local AI:* The prototype calls third-party model APIs for chat, deep research, and multimodal embeddings, which is not acceptable for sensitive OSINT workloads. A full deployment will run open-source LLMs and the multimodal embedding generator on a dedicated multi-GPU server, giving control over latency, cost per query, data custody, and model selection while supporting RAG-based chat, semantic search, and similarity search across the Intel store.

The implementation plan and budget that follow estimate these requirements as closely as possible at this stage. The figures are subject to revision as implementation, testing, and the beta deployment reveal new requirements and real operating conditions.

== Implementation Timeline
The implementation effort will focus on hardening, scaling, extending, and operationalising the validated architecture.

*Stage 1: Testing and Troubleshooting* (4 Months)

+ *Server Setup:* Setup of 2 lightweight nodes in the microservice fleet to test the fetchers under development.
+ *Microservice development and testing:* Build and test microservices for Reddit, Facebook, X, Instagram, websites, LinkedIn, GDELT, YouTube, BlueSky, OpenSky, AIStream, and ActivityPub.
+ *Intel schema finalisation:* Take reviews from concerned personnel to finalise the Intel schema for the task at hand.
+ *Reports and alerts:* Set up the reports and alerts system.
+ *Core product finalisation:* Finalise the core product based on feedback.
+ *Standard operating procedures:* Formalise SOPs for maintenance and upgrade of nodes and microservices.

*Stage 2: Pre-Scaling* (4 Months)

+ *Local embedding model and LLM:* Deploy the embedding model and LLM on a local machine.
+ *Beta testing:* Beta test with a limited user base.
+ *Amendments:* Ship changes driven by those tests.

*Stage 3: Scaling Up* (4 Months)

+ *Multi-region nodes:* Large-scale deployment of nodes across multiple regions (node counts to be decided from test results).
+ *On-demand microservices:* Add scraping microservices on demand.
+ *Complete rollout:* Full rollout to the intended user base.

#figure(
  table(
    columns: (1fr, auto, auto),
    align: (left + horizon, center + horizon, center + horizon),
    stroke: 0.5pt,
    fill: (_, y) => if y == 0 { gray.lighten(75%) },
    table.header[*Activity*][*Stage*][*Estimated duration*],
    [Server setup], [Stage 1], [2 weeks],
    [Microservice development and testing (12 sources)], [Stage 1], [14 weeks],
    [Intel schema finalisation], [Stage 1], [4 weeks],
    [Reports and alerts system], [Stage 1], [6 weeks],
    [Core product finalisation], [Stage 1], [4 weeks],
    [Standard operating procedures], [Stage 1], [3 weeks],
    [Local embedding model and LLM deployment], [Stage 2], [5 weeks],
    [Beta testing with limited user base], [Stage 2], [10 weeks],
    [Amendments based on beta feedback], [Stage 2], [5 weeks],
    [Multi-region node deployment], [Stage 3], [8 weeks],
    [On-demand microservice additions], [Stage 3], [6 weeks],
    [Complete rollout], [Stage 3], [6 weeks],
  ),
  caption: [Per-activity duration estimates. Activities within a stage run in parallel; each stage spans approximately 4 months, for a 12-month total programme.],
)



== Estimated Budget

The figures below are an indicative one-year budget for the full-scale deployment described above. Node and infrastructure counts are placeholders that will be revised once Stage 1 test results are in.

#figure(
  table(
    columns: (1fr, auto),
    align: (left + horizon, right + horizon),
    stroke: 0.5pt,
    fill: (_, y) => if y == 0 { gray.lighten(75%) },
    table.header[*Item*][*Estimate (INR)*],
    [Team of 10 developers to maintain fetcher microservices (1 year)], [60 Lakh],
    [50-70 microservice nodes (estimate, to be revised per plan)], [25 Lakh],
    [Internet and operating costs for microservice nodes], [2.5 Lakh],
    [Nvidia GB300 server for local AI deployment], [80 Lakh],
    table.cell(fill: gray.lighten(85%))[*Total*],
    table.cell(fill: gray.lighten(85%))[*167.5 Lakh*],
  ),
  caption: [Estimated one-year budget for full-scale deployment of Seer.],
)



#pagebreak()

= Challenges & Mitigation

A platform of this scope faces challenges that are well understood from the start, and the architecture in the earlier sections is designed to absorb them rather than treat them as afterthoughts. The subsections below pair each principal challenge with the specific mitigations already built into Seer's design or planned as part of the full-scale deployment.

== Source Volatility and Active Blocking

Public sources are the most unreliable input to any OSINT system. Layouts change, identifiers move, request patterns are throttled, IP ranges are flagged, behavioural signals are scored, and CAPTCHAs are introduced without notice. A scraper that works today can break next week, and a stable address can be banned the moment a platform updates its defences.

Seer responds at three layers. Each fetcher is an independent microservice, so a broken or banned fetcher does not stop the rest of the platform; only that one source's collection is interrupted while the team repairs it. The fetcher fleet is then distributed across multiple regions, networks, and hosting providers so that a block on a single node or network does not interrupt collection from that source. Within each fetcher, source-specific anti-blocking mechanisms like proxy rotation, browser spoofing, fingerprint randomisation, CAPTCHA handling, and human-behaviour emulation are treated as part of the plugin rather than a global toggle, so each platform can be matched with the strategy that actually works for it.

== Noise, Duplication, and Data Quality

Public web data carries far more irrelevant content than useful intelligence. Entertainment, advertising, repeat posts, and unrelated chatter outweigh the signal in almost every feed. If this material reaches the Intel layer unfiltered, search results, reports, alerts, and AI responses all degrade together.

Seer pushes filtering up the pipeline rather than relying on downstream cleanup. Each Source Plugin runs source-specific pre-processing, deduplication against its own raw store, and topic-aware filtering that supports both natural-language objectives — for example, "reports of military movement near a border region" — and scripted rules for technical users. Raw, source-native records are retained alongside the cleaned Intel so that a record discarded in error can always be recovered, and so any later question about what was filtered has a clear answer.

== Misinformation and Source Credibility

OSINT collection cannot assume sources are honest. Some platforms host disinformation by design, others are compromised, and many simply repeat low-confidence claims. A platform that flattens all sources into one search index will surface contradictions without context, and a single bad source can pull an analyst toward the wrong conclusion.

Every Intel record in Seer keeps a provenance link to the raw source record that produced it, including the source plugin, the fetch time, and the source-native identifiers. The Intel Layer can therefore compare claims across sources, flag contradictions, and let analysts trace each piece of Intel back to its origin. Repeated conflict from a particular source can be tracked over time, turning provenance from passive auditing into a credibility signal the analyst can act on rather than a feature buried in the database.

== AI Accuracy and Grounded Outputs

LLMs are useful for structured extraction, summarisation, semantic search, and analyst chat, but they can hallucinate, generalise inappropriately, or invent details. Untreated, this risks producing intelligence that is fluent but incorrect, which is worse than no answer at all in operational use.

Seer keeps generative AI grounded in stored evidence. Chat and report workflows use Retrieval Augmented Generation against the Intel Layer, so answers are anchored to retrieved Intel records and, through them, to raw source material. Outputs surface their evidence links rather than producing untraceable narrative, and analyst review remains in the workflow for any decision that matters. Running open-source LLMs on the dedicated GPU server during full-scale deployment further avoids surrendering custody of sensitive prompts and retrieved Intel to a third-party API.

== Scalability and Operational Load

At full scale, scraping, embedding generation, storage, and AI processing all become resource-intensive at the same time. A naive monolithic deployment would couple their failure modes: a slow embedding job would back up scraping, and a single overloaded host would degrade everything downstream of it.

The architecture separates these concerns so they can scale independently. The fetcher fleet runs as horizontally scaled microservices; embedding and AI workloads run on the dedicated multi-GPU server provided for in the budget; raw storage is per source so that a busy feed cannot starve quieter ones; and queues, retries, and backpressure are handled at the boundaries between stages. Operational visibility — job failure rates, queue depth, fetcher health, and service-level metrics — is treated as a first-class deployment requirement rather than a later addition, and the SOPs formalised in Stage 1 are the artefact that keeps that visibility actionable.

== Security, Access Control, and Data Custody

A system that monitors sensitive subjects produces sensitive outputs. Operational topics, watchlists, alert conditions, and Intel records cannot be exposed to general application users, and the maintenance access needed to repair fetchers must not extend to analyst data.

Seer is built on least-privilege boundaries. Fetcher microservices are operationally separate from the analyst-facing application, so a developer maintaining a Reddit or website scraper does not require, and does not get, access to analyst workflows or Intel. The deployment plan includes role-based access control, audit logs, encrypted transport, and a clean separation between collection services, databases, and the analyst application. Hosting LLMs and the multimodal embedding model on the on-premise GPU server removes the remaining need to send sensitive prompts or retrieved Intel to external providers.

== Summary

None of the challenges above are eliminated permanently — they are kept localised, traceable, and operationally manageable. Source breakage stays inside one plugin; noise is removed before it pollutes Intel; conflicting claims are surfaced with their provenance; AI outputs stay tied to evidence; load is absorbed by independently scaled services; and sensitive data does not leave the boundaries of the deployment. Together with the implementation timeline and budget in the previous section, these mitigations make Seer realistic to operate, not only realistic to demonstrate.
