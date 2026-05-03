#import "@preview/fletcher:0.5.8" as fletcher: diagram, node, edge

#set page(margin: 0.69in)
#set text(size: 12pt)
#set par(justify: true, leading: 0.65em)
#show heading: it => block(above: 1em, below: 0.69em, it)

#outline()

#pagebreak()

#align(center)[
  #text(size: 22pt, weight: "bold")[Proposed Technical Solution]
]

#text(style: "italic")[You can preview a prototype of the proposed solution - Seer at https://seer.lariv.in/.]

Seer is proposed as a distributed, large scale OSINT analysis and monitoring platform. The system is designed to collect large volumes of public and user-provided information through source plugins, store it in a common Intel layer, process it through AI-assisted analysis plugins, and present actionable outputs through a managed role-based application. It directly addresses the problem statement by reducing manual collection effort, enabling real-time monitoring of user-defined subjects, and improving the reliability, traceability, and usability of intelligence outputs.

The proposed system is not starting from a purely theoretical design. A working prototype of Seer already exists and demonstrates the core architecture: a Reddit source, website scraper, deep research AI agent, AI chat interface, Intel layer, multimodal embedding support, Retrieval Augmented Generation (RAG), scheduled scraping workers, and a GIS map. The grant would therefore be used to harden and scale an existing architecture rather than to begin exploratory development from zero.

The current prototype validates the core direction of the platform. The proposed work will focus on preparing it for larger and more demanding use: scaling collection across more sources, improving source-specific plugins, strengthening the Intel layer for military requirements, adding stronger reliability and misinformation checks, and preparing for deployment needs such as rate limits, unstable sources, distributed workers, access control, and audit logs.


= 1. Technical Architecture and Approach

Seer follows a layered, plugin-based architecture. A fleet of data-scraping microservices collects public web data. Other services, including public datasets, API-accessible sources, and proprietary services, can feed the platform through the same source-plugin interface. 

Each source is represented by an independent plugin, so new source types can be added without changing the rest of the system.
Within each source plugin, collected material passes through a common pipeline for data fetching, pre-processing, filtering, raw data storage, and extraction of content suitable for intelligence generation. Source-specific stores preserve the original or source-native records.

The Intel layer processes the extracted content into a defined format and feeds it into a shared model before saving it to the database. Downstream applications then consume this Intel model for semantic search, AI chat, similarity search, report generation, interactive maps, and alerts and warnings, while retaining links back to the underlying source records where available.

#figure(
  image("Seer Architecture.svg", height: 72%),
  caption: [Seer high-level system architecture],
)

== 1.1 Data Scraping Microservice Fleet

The first layer of Seer is a fleet of independent data-scraping microservices. Instead of placing all scraping logic inside the main application, each major web source can have its own dedicated fetcher, such as a Reddit scraper, website scraper, Twitter scraper, or future source-specific scraper. This keeps the data collection layer operationally separate from the core Seer application. If one fetcher is rate-limited, blocked, banned, or broken because a website changes its layout, the rest of the platform continues to function.

This microservice architecture also allows Seer to implement redundancy at the correct level. Public web sources frequently use bot-detection systems, rate limits, suspicious traffic detection, fingerprinting, CAPTCHA checks, and behavioural analysis. Seer can respond to these constraints using proxy rotation, browser spoofing, fingerprint randomisation, CAPTCHA solving, and human-behaviour emulation. These capabilities can be shared across the fleet while still allowing source-specific handling. For example, the safeguards needed for a social media platform may be different from the safeguards needed for a news website, forum, or public search page.

Separating scrapers into independent services also improves scale and maintainability. Large-scale fetching can run in parallel across many workers without overloading the main application. Scrapers can be deployed, restarted, scaled, or replaced independently. Different developer teams can maintain different fetchers without needing full access to the main application, unrelated source plugins, analyst workflows, or sensitive operational data. This reduces security and privacy risk while allowing the collection layer to evolve quickly as external websites and platforms change.

== 1.2 Other Services

The architecture is not limited to web scraping. Seer can also ingest information from other programmatic services, including public datasets, API-accessible sources, proprietary services, and future private data feeds. These services enter the same source-plugin workflow as scraper-backed sources, which means the platform can treat a structured API, an open government dataset, an aviation feed, a maritime feed, or an internal proprietary source as another source of intelligence.

This design gives Seer a single ingestion model for many kinds of data. A source may be fetched by a scraper microservice, queried through an authenticated API, received from a public dataset, or connected through a proprietary integration, but once it reaches the source-plugin layer it can be processed, filtered, stored, and converted into Intel using the same architectural pattern. This makes it possible to add new sources without redesigning downstream search, reporting, maps, AI chat, or alerting features. It also allows Seer to run multiple data fetchers for a single source type, ensuring coverage for sources that require strict full-time monitoring even if one fetcher goes down, is rate-limited, or temporarily loses access.

== 1.3 Source Plugins

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

== 1.4 Intel Layer

The Intel Layer is responsible for converting the source-specific material produced by Source Plugins into a common intelligence model. Source Plugins may produce raw posts, webpages, comments, event records, API responses, media files, and other source-native objects. The Intel Layer extracts the most vital information from this material and reformats it into a standard Intel record that downstream applications can search, compare, analyse, and visualise.

This extraction process uses LLMs to identify and structure the important parts of the raw content. Depending on the source, the Intel Layer can extract a title, description, datetime, source kind, relevant entities, locations, events, summaries, claims, and other metadata needed for analysis. The goal is not merely to store raw content, but to convert it into a consistent intelligence object that can be used reliably across the platform.

Every Intel record preserves a link back to the original raw data stored by the Source Plugin. This creates a clear origin trail for each piece of intelligence: what source produced it, when it was fetched, what raw record it came from, and how it entered the Intel database. This traceability is essential for audit, verification, and analyst confidence.

The same origin trail also supports misinformation and source-integrity analysis. If two sources feed Seer conflicting Intel, analysts can trace both claims back to their raw records and source histories. Over time, this helps identify unreliable, compromised, or deliberately misleading sources. Instead of treating Intel as isolated text, Seer can evaluate where it came from and whether the source has a pattern of producing conflicting or low-confidence information.

After the Intel record is structured, Seer generates vector embeddings for the content. For multimodal material, including text, images, and other media where available, Seer can use an open-source multimodal embedding model to represent the content in vector space. These embeddings power downstream applications such as semantic search, similarity search, retrieval for AI chat, clustering, correlation, and map or report workflows.

== 1.5 Downstream Applications

The architecture above enables several downstream applications to operate on the same Intel model. Because content from different sources is normalised into a shared structure and embedded into a common vector space, Seer can search and analyse information across websites, social platforms, public datasets, APIs, and proprietary feeds at the same time.

Semantic search allows analysts to search the Intel database by meaning rather than only by exact keywords. Similarity search uses Intel embeddings to find records that are semantically or visually close to a selected item. This helps surface related posts, articles, events, images, or reports even when they use different wording or come from different source types.

The shared Intel structure also allows Seer to identify connections across sources and across time. For example, a social media post, a news article, a public event record, and an aviation or maritime signal may each describe part of the same activity. By analysing them together, Seer can help reveal patterns that would be difficult to detect inside a single feed. This is especially valuable for monitoring evolving situations where weak signals appear gradually across multiple sources.

Seer also includes AI tools such as an analyst chatbot that can query the Intel database and answer questions using the collected evidence. The chatbot can be equipped to trigger new source fetches, inspect gaps in the available intelligence, and add newly discovered sources into the monitoring workflow. This turns the assistant from a passive question-answering tool into an active interface for expanding collection.

A report generator has also been implemented. It can query the Intel gathered about a requested topic and produce a detailed, comprehensive report suitable for review by officials. Because the report is generated from structured Intel records linked to raw source material, the output can retain evidence trails rather than becoming an untraceable narrative summary.

Seer can also support alert and warning systems. Analysts can define alert conditions in natural language, such as the appearance of a specific threat pattern, location, source combination, or operational signal. Whenever newly added Intel matches those conditions, the alert can trigger automatically. For users who require more control, the same alerting system can support scripted conditions.

Geospatial visualisation is another downstream application. Intel with location metadata can be plotted onto an interactive map to help analysts identify spatial patterns, clusters, movement, proximity, and regional changes over time. This can support situational awareness, field planning, and the discovery of geographic relationships between otherwise disconnected pieces of Intel.

The same architecture can support additional applications such as person-of-interest tracking, knowledge graphs, source reliability dashboards, timeline reconstruction, entity correlation, and automated briefings. The important point is that Seer's architecture does not lock the system into a single dashboard or workflow. Once the military's requirements, operational procedures, and existing systems are studied in detail, the same source-plugin and Intel-layer foundation can be used to build the downstream tools that best fit the mission.




= 2. Innovation

Seer's innovation lies in how it combines resilient source collection, source-aware AI processing, provenance-based intelligence analysis, and agentic downstream workflows into one extensible OSINT architecture. The following capabilities distinguish it from a basic OSINT dashboard or a conventional scraping-and-search system.

== 2.1 SourceWeave: Multi-modal and multi-source data processing

SourceWeave is Seer's cross-plugin processing mechanism. It allows one Source Plugin to invoke selected functions from another Source Plugin when the incoming data contains linked, embedded, or dependent material. For example, if the Reddit Source Plugin encounters a URL inside a post or comment, it can call the Website Source Plugin to fetch, parse, and process that page before feeding both pieces of evidence into the Intel Layer.

This is more powerful than treating each source as an isolated feed. Real OSINT data is often nested across platforms: a post links to a website, a website references a social account, a profile links to media, and a dataset points to external records. SourceWeave allows Seer to follow these relationships through controlled plugin-to-plugin calls, enabling multi-source and multi-modal intelligence extraction without duplicating parsing logic across plugins.

== 2.2 Adaptive Fetcher Camouflage

Different platforms use different anti-bot and abuse-detection systems. Seer therefore treats anti-blocking logic as a source-specific capability rather than a generic scraper setting. Fetchers can use proxy rotation, browser spoofing, fingerprint randomisation, CAPTCHA handling, and human-behaviour emulation in ways tailored to the source being monitored.

This allows Seer to adapt to the defensive behaviour of each target platform. A news website, social media platform, forum, public search page, or structured dataset may each require a different collection strategy. By isolating these strategies inside fetcher microservices, Seer can harden collection without exposing or destabilising the rest of the system.

== 2.3 Language based data fetching

Seer allows users to define what they are looking for at the source level, before irrelevant material becomes Intel. Analysts can describe the desired information in natural language, while advanced users can provide scripts for stricter filtering requirements. This hybrid filtering model helps remove advertisements, unrelated articles, broad search noise, and irrelevant social media content before it enters the Intel database.

The innovation is that filtering happens as part of the source pipeline, not only after data has already been collected and indexed. This reduces noise in downstream search, reports, alerts, and AI responses, while still giving technical users the ability to enforce precise rules when required.

== 2.4 Source credibility tracking

Every Intel record remains linked to the raw source record that produced it. Beyond auditability, this enables Seer to compare conflicting claims by tracing them back to their original sources. If multiple sources repeatedly produce contradictory or low-confidence Intel, Seer can help analysts identify unreliable feeds, compromised sources, or coordinated misinformation patterns.

This turns provenance from simple record-keeping into an analytical capability. The system can support source credibility assessment by examining not only what was claimed, but where the claim originated and whether that source has a history of conflict with other evidence.

== 2.5 Context-Aware Intel Extraction

Seer's Intel extraction is not a one-size-fits-all summarisation step. Each source type can provide its own extraction context before the Intel Layer structures the record using LLMs. A Reddit post may require comments, author identity, subreddit context, and engagement metadata. A webpage may require canonical URL detection, article extraction, publication dates, and linked media. A professional-network profile may require connections and organisational context.

This makes the resulting Intel model richer than a generic scraped text record. The system preserves source-specific meaning while still converting the output into a common intelligence format.

== 2.6 Multimodal Intel Embedding

Seer can represent text, images, and other supported media in a shared vector space using open-source multimodal embedding models. This allows downstream applications to search and compare intelligence by semantic and visual similarity, not just by keywords or source type.

For OSINT workflows, this is useful when related evidence appears in different forms: an image, a short post, a long article, and a structured dataset may all refer to the same event. Multimodal embeddings allow these signals to be discovered together.

== 2.7 Agentic Collection Interface

Seer's AI chatbot is designed to be more than a passive question-answering interface. It can query the Intel database, answer analyst questions with evidence, identify gaps in available information, trigger new source fetches, and add newly discovered sources into the monitoring workflow.

This creates a more active analyst workflow. Instead of manually moving between search, source configuration, collection, and reporting screens, users can ask the system to investigate, expand collection, and return evidence-backed answers through one interface.

== 2.8 Natural-Language Alerting

Seer can support alert and warning conditions written in natural language. Analysts can describe a condition of interest, such as a threat pattern, location-specific activity, source combination, or emerging operational signal. When newly added Intel matches that condition, the system can trigger an alert.

This reduces the effort required to create monitoring rules for complex, evolving situations. The same alerting layer can also support scripted rules where strict deterministic behaviour is required.

== 2.9 Mission-Adaptive Intelligence Applications

The architecture is intentionally open-ended at the downstream layer. Because all sources feed into a common Intel model, Seer can support additional military-specific applications such as person-of-interest tracking, knowledge graphs, source reliability dashboards, timeline reconstruction, entity correlation, spatial pattern analysis, and automated briefings.

This means the platform is not limited to a fixed set of dashboards. After studying the military's requirements, existing systems, and operational procedures, Seer can extend the same core architecture into the applications that provide the highest mission value.

= Implementation and Feasibility
The proposed implementation is feasible because Seer is not a concept-only proposal. A *Working Prototype* is already available at https://seer.lariv.in/ and demonstrates the core platform capabilities needed for this solution: source plugins, Reddit and website collection, scheduled scraping workers, an Intel layer, multimodal embedding support, RAG-based AI chat, deep research workflows, report generation, and GIS visualisation. The implementation effort will therefore focus on hardening, scaling, extending, and operationalising a validated architecture.

Implementation will proceed by strengthening the existing prototype while keeping the analyst-facing capabilities active from the beginning.

In the first stage:
+ *Workers:* improve reliability of scheduled collection and background jobs.
+ *Intel schema:* formalise the shared intelligence model so downstream tools share one structure.
+ *Logging and audit trails:* strengthen traceability and operational review.
+ *Scraping microservice fleet:* complete the pattern of separate fetcher services per major source, outside the main application, each deployable and scalable on its own—building on the Reddit and website workers already shown in the prototype.

In the second stage:
+ *RAG-based AI chat:* harden chat that answers from stored Intel with clear evidence links.
+ *Semantic and similarity search:* mature meaning-based retrieval and “find alike” workflows over Intel.
+ *Report generation:* stabilise automated reports suitable for review.
+ *Alerting:* harden natural-language and scripted alert conditions against incoming Intel.
+ *Map visualisation:* mature GIS-style views tied to geospatial Intel.

In the third stage:
+ *New source plugins:* add plugins for additional sources using the same model as the existing Reddit and website collectors (scheduling, credentials and rate limits, raw retention, and mapping into Intel stay encapsulated per plugin).
+ *Controlled rollout for feedback:* deploy to a deliberately limited set of users so structured feedback reflects real workflows before access is widened.
+ *Amendments from feedback:* prioritise and ship changes driven by that feedback—fixes, UX adjustments, and source or workflow tuning—before scaling the user base.

#figure(
  image("implementation.svg", width: 69%),
  caption: [Seer phased implementation roadmap.],
)

Misinformation and conflicting stories are addressed on top of the common Intel model. Intel entries stay linked to the original captures (for example web address, time, and a fingerprint of the content where that helps). The system can flag tensions between sources by comparing who, what, where, and when; when analysts correct or dismiss a finding, that is logged so later improvements apply across all feeds, not just one.

Collection is split into separate services and background jobs. Raw material is stored per source, while the curated Intel and search layers are shared. If one site changes its layout, blocks access, or throttles traffic, work on other sources can continue, and extra capacity can be added for busy feeds without redesigning the analyst application.

The same separation supports integration with existing processes. The web app, reports, alerts, and maps all draw from the same Intel; in time, other systems could connect through secure interfaces and exports that keep the link from a conclusion back to the underlying source material. That way important decisions can be checked against evidence, not only against generated text.

The *Working Prototype* already covers the difficult joins between plugins, background collection, multimodal search, and chat grounded in stored Intel. What remains is steady engineering work: evolving the data model without breaking existing users, clearer operational monitoring, access control suitable for sensitive topics, sensible handling of busy queues and retries, and proving the platform under larger volumes of sources and traffic.

= Challenges & Mitigation (if any)
The main implementation challenges are source reliability, data quality, AI accuracy, scalability, and secure deployment. Public web sources may change layouts, restrict access, rate-limit traffic, or block automated collection. Seer mitigates this through independent source plugins, separate fetcher microservices, scheduled retries, source-specific handling, and the ability to update or replace one scraper without affecting the rest of the platform.

OSINT data can also contain noise, duplication, misinformation, and conflicting claims. Seer addresses this through pre-processing, deduplication, source-level filtering, raw data preservation, provenance tracking, and cross-source comparison. Analysts will be able to trace extracted Intel back to the original source record before acting on important outputs.

AI-generated analysis may occasionally produce incomplete or incorrect interpretations. This will be mitigated by using RAG over stored Intel, preserving evidence links, adding confidence and source-context indicators where appropriate, and keeping analyst review in the workflow for critical decisions.

At larger scale, scraping, embedding generation, storage, and AI processing can become resource-intensive. The architecture mitigates this through distributed workers, horizontal scaling, queue-based processing, source-specific raw stores, and monitoring of job failures, queue depth, and service health.

Security and access control are also important because the system may handle sensitive user-defined monitoring topics and operational outputs. The deployment plan will therefore include role-based access control, audit logs, encrypted transport, controlled user access, and separation between collection services, databases, and analyst-facing applications.
