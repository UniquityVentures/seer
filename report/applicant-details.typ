#set heading(numbering: "A.")

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
  #text(size: 22pt, weight: "bold")[Applicant Details]
]

#v(0.5em)

= Applicant's Resume

*Name:* Raj Tagore \
*Role:* Founder, Uniquity Ventures \
*Website:* https://rajtagore.com/

#v(0.5em)

*Education*

*MSc Robotics — King's College London* #h(1fr) *Sep 2023 – Oct 2024*

- *Grade:* Distinction
- *Key modules:* AI, ML, Multi-Agent Systems

#v(0.5em)

*Experience*

*Founder — Uniquity Ventures* #h(1fr) *Oct 2024 – Present*

- Working on building deep-tech products in robotics and AI0
- Researching and building AI and ERP systems for the past year under the *Lariv* brand

#v(0.5em)

= Relevant Information related to solution

*Organisation:* Uniquity Ventures (MSME, registered July 2023) \
*Brand:* Lariv (https://lariv.in/)

For the past year, Uniquity Ventures has been researching, building, and selling ERP systems under the *Lariv* brand. Through Lariv, the team has built direct experience with concerns that recur in the proposed OSINT system: ingestion from heterogeneous data sources, durable storage and indexing, role-based access control, scheduled background workers, AI-assisted workflows, and the reliability requirements of running a service that paying customers depend on. Several of these architectural foundations carry over directly to Seer.

The applicant's *MSc in Robotics* with a Distinction grade and key modules in *Artificial Intelligence* and *Multi-Agent Systems* provides formal grounding in the AI techniques and multi-agent architectures that underpin Seer's source-plugin design, retrieval-augmented generation, multimodal embedding, and agentic analyst tooling.