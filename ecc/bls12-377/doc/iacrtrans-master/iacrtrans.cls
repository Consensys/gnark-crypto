\def\fileversion{0.92}
\def\filedate{2020/08/24}

\NeedsTeXFormat{LaTeX2e}[1995/12/01]
\typeout{^^J  ***  LaTeX class for IACR Transactions v\fileversion\space ***^^J}
\ProvidesClass{iacrtrans}[\filedate]

% IACR Transactions DOCUMENT CLASS
% Written by Gaetan Leurent gaetan.leurent@inria.fr and others (2016-2020)
%
% To the extent possible under law, the author(s) have dedicated all
% copyright and related and neighboring rights to this software to the
% public domain worldwide. This software is distributed without any
% warranty.
%
% You should have received a copy of the CC0 Public Domain Dedication
% along with this software. If not, see
% <http://creativecommons.org/publicdomain/zero/1.0/>.
%
%
%%% Class options:
%
%% Document mode
% [preprint]      Preprint (no copyright info) -- default mode
% [submission]    Anonymous submission
% [notanonymous]  Keep author names in submission mode
% [final]         Final version
% [journal=tosc]
% [journal=tches]
% [draft]
%% Package options
% [spthm]         Emulate llncs sptheorem and remove automatic \qed in proof
% [floatrow]      Load floatrow package with correct captions
% [nohyperref]    Disable automatic loading of hyperref
% [nolastpage]    Disable automatic loading of lastpage
% [xcolor=xxx]    Pass xxx to xcolor package
% [hyperref=xxx]  Pass xxx to hyperref package
%
%%% HOWTO use this class
%
%% Title
% \title[short]{Long title}
%
%% Authors/affiliation:
% \author{Alice \and Bob}
% \institute{ABC\\ \email{alice@abc} \and DEF\\ \email{bob@def}}
%
%% Keywords/abstract:
% \keywords{banana \and apple}
% \begin{abstract}
% Lorem ipsum dolor sit amet...
% \end{abstract}
%
%% Warnings
% - please don't use any \pagestyle or \thispagestyle command
% - if you have proof with explicit \qed inside, you should either
%   remove \qed symbols, replace them by \qedhere, or add option [spthm]


% Common definitions
\RequirePackage{xkeyval}
\def\publname{IACR Transactions}
\def\publnameshort{IACR Transactions}
\define@choicekey*+{IACR}{journal}[\val\nr]{tosc,tches}{%
    \ifcase\nr\relax
        \def\publname{IACR Transactions on Symmetric Cryptology}%
        \def\publnameshort{IACR ToSC}%
    \or
        \def\publname{IACR Transactions on Cryptographic Hardware and Embedded Systems}%
        \def\publnameshort{IACR TCHES}%
    \fi
}{%
    \ClassError{iacrtrans}{journal value is only allowed to be: tosc, or tches}%
}
\def\IACR@vol{0}
\def\IACR@no{0}
\def\IACR@fp{1}
\def\IACR@lp{\if@loadhr\pageref*{LastPage}\else\pageref{LastPage}\fi}
\def\IACR@ISSN{XXXX-XXXX}
\def\IACR@DOI{XXXXXXXX}
\def\IACR@Received{20XX-XX-XX}
\def\IACR@Revised{20XX-XX-XX}
\def\IACR@Accepted{20XX-XX-XX}
\def\IACR@Published{20XX-XX-XX}
\newif\if@IACR@Received \@IACR@Receivedfalse
\newif\if@IACR@Revised \@IACR@Revisedfalse
\newif\if@IACR@Accepted \@IACR@Acceptedfalse
\newif\if@IACR@Published \@IACR@Publishedfalse

\newcommand{\setfirstpage}[1]{\def\IACR@fp{#1}\setcounter{page}{#1}}
\newcommand{\setlastpage}[1]{\def\IACR@lp{#1}}
\newcommand{\setvolume}[1]{\def\IACR@vol{#1}}
\newcommand{\setnumber}[1]{\def\IACR@no{#1}}
\newcommand{\setISSN}[1]{\def\IACR@ISSN{#1}}
\newcommand{\setDOI}[1]{\def\IACR@DOI{#1}}

\newcommand{\setReceived}[1]{\@IACR@Receivedtrue\def\IACR@Received{#1}}
\newcommand{\setRevised}[1]{\@IACR@Revisedtrue\def\IACR@Revised{#1}}
\newcommand{\setAccepted}[1]{\@IACR@Acceptedtrue\def\IACR@Accepted{#1}}
\newcommand{\setPublished}[1]{\@IACR@Publishedtrue\def\IACR@Published{#1}}

% Options
\newif\if@loadhr
\@loadhrtrue
\newif\if@hyperxmp@doi
\@hyperxmp@doifalse
\define@key{IACR}{nohyperref}[]{\@loadhrfalse}
\newif\if@floatrow
\@floatrowfalse
\define@key{IACR}{floatrow}[]{\@floatrowtrue}
\newif\if@submission
\@submissionfalse
\newif\if@anonymous
\@anonymousfalse
\newif\if@preprint
\@preprinttrue
\define@key{IACR}{final}[]{\PassOptionsToClass{\CurrentOption}{article}\@preprintfalse}
\define@key{IACR}{preprint}[]{\@preprinttrue}      % Default
\define@key{IACR}{submission}[]{\@submissiontrue\@anonymoustrue}
\define@key{IACR}{draft}[]{\@preprinttrue\PassOptionsToClass{\CurrentOption}{article}}
\define@key{IACR}{notanonymous}[]{\@anonymousfalse}
\newif\if@spthm
\@spthmfalse
\define@key{IACR}{spthm}[]{\@spthmtrue}
\newif\if@IACR@lastpage
\@IACR@lastpagetrue
\define@key{IACR}{nolastpage}[]{\@IACR@lastpagefalse}

\define@key{IACR}{xcolor}{\PassOptionsToPackage{#1}{xcolor}}
\define@key{IACR}{hyperref}{\PassOptionsToPackage{#1}{hyperref}}

\DeclareOptionX*{\PassOptionsToClass{\CurrentOption}{article}}
\ProcessOptionsX<IACR>\relax

% article class with a4paper
\LoadClass[10pt,twoside]{article}[2007/10/19]

% Hyperref must be loaded last, but before other AtEndPreamble hooks
\if@loadhr
  \RequirePackage{xcolor}
  \RequirePackage{etoolbox}
  \AtEndPreamble{
    \@ifpackageloaded{hyperref}{}{\RequirePackage{hyperref}}
    \hypersetup{pdflang=en}
    \hypersetup{colorlinks=true,
      citecolor=black!70!green,
      linkcolor=black!70!red}
    % Disable latexdiff commands in PDF links
    \pdfstringdefDisableCommands{%
      \def\DIFadd#1{#1}%
      \def\DIFdel#1{}%
    }
    % Old versions of hyperxmp do not support DOI
    \@ifpackagelater{hyperxmp}{2019/03/14}{\@hyperxmp@doitrue\hypersetup{keeppdfinfo=true}}{}
  }
  \RequirePackage{hyperxmp} % hyperxmp 5.x has issues with AtEndPreamble
  \setcounter{tocdepth}{2}
\fi


% Geometry
\RequirePackage[a4paper,hscale=0.65,vscale=0.75,marginratio=1:1,marginparwidth=2.7cm]{geometry}
\RequirePackage{afterpage}
% Title fonts: bf+sf
\RequirePackage{sectsty}
\allsectionsfont{\sffamily\boldmath}
% Also for descriptions
\renewcommand*\descriptionlabel[1]{\hspace\labelsep
                                   \normalfont\bfseries\sffamily\boldmath #1}


% Title/Author/affiliations
\def\@institute{No institute given.}
\newcommand{\institute}[1]{\gdef\@institute{#1}}
\newcommand{\authorrunning}[1]{\gdef\IACR@runningauthors{#1}}
\newcommand{\titlerunning}[1]{\gdef\IACR@runningtitle{#1}}

\newcounter{IACR@author@cnt}
\newcounter{IACR@inst@cnt}
\newif\if@IACR@autoinst
\@IACR@autoinsttrue
\def\IACR@author@last{0}

\renewcommand\maketitle{\par
  \begingroup
    \renewcommand\thefootnote{\@fnsymbol\c@footnote}%
    \long\def\@makefntext##1{\parindent 1em\noindent
            \hb@xt@1.8em{%
                \hss\@textsuperscript{\normalfont\@thefnmark}}##1}%
    \newpage
    \global\@topnum\z@   % Prevents figures from going at top of page.
    \@maketitle
    \thispagestyle{title}\@thanks
  \endgroup
  \setcounter{footnote}{0}%
  \global\let\thanks\relax
  \global\let\maketitle\relax
  \global\let\@maketitle\relax
  \global\let\@thanks\@empty
%  \global\let\@author\@empty
  \global\let\@date\@empty
%  \global\let\@title\@empty
  \global\let\title\relax
  \global\let\author\relax
  \global\let\date\relax
  \global\let\and\relax
  % Adjust header size for title page
  \addtolength{\headheight}{\baselineskip}%
  \addtolength{\headsep}{-\baselineskip}%
  \afterpage{%
    \global\advance\headheight by -\baselineskip%
    \global\advance\headsep by \baselineskip%
  }%
}
\def\@maketitle{%
  % Count authors and affiliations
  \setcounter{IACR@author@cnt}{1}%
  \setcounter{IACR@inst@cnt}{1}%
  \setbox0\hbox{\def\thanks##1{\global\@IACR@autoinstfalse}\def\inst##1{\global\@IACR@autoinstfalse}\def\and{\stepcounter{IACR@author@cnt}}\@author}%
  \setbox0\hbox{\def\and{\stepcounter{IACR@inst@cnt}}\@institute}%
  \xdef\IACR@author@last{\theIACR@author@cnt}%
  \edef\IACR@inst@last{\theIACR@inst@cnt}%
  \ifnum\IACR@author@last=\IACR@inst@last\else\@IACR@autoinstfalse\fi
  \ifnum\IACR@author@last=1 \@IACR@autoinstfalse\fi
  \newpage
  \null
  \vskip 2em%
  \begin{center}%
  \let \footnote \thanks
    {\def\@makefnmark{\rlap{\@textsuperscript{\normalfont\@thefnmark}}}%
      {\LARGE \bfseries\sffamily\boldmath \@title\par}
    \ifdefined\@subtitle\vskip .5em{\large\sffamily\bfseries\@subtitle\par}\fi}%
    \vskip 1.5em%
    {\large
      \lineskip .5em%
        \if@anonymous
        Anonymous Submission
        \else
        \setcounter{IACR@author@cnt}{1}%
        \def\and{\if@IACR@autoinst\inst{\theIACR@author@cnt} \fi
          \stepcounter{IACR@author@cnt}%
          \ifnum\theIACR@author@cnt=\IACR@author@last\unskip\space and \ignorespaces\else\unskip, \ignorespaces\fi}
        \@author\if@IACR@autoinst\inst{\theIACR@author@cnt}\fi
        \vskip 1em\par
        \small
        \setcounter{IACR@author@cnt}{1}%
        \def\and{\par\stepcounter{IACR@author@cnt}$^\theIACR@author@cnt$~}
        \ifnum\IACR@inst@last>1 $^1$~\fi\ignorespaces%
        \@institute
        \fi
      }%
  \end{center}%
  \par
  \vskip 1.5em}

\def\author{\@ifnextchar[{\IACR@@@author}{\IACR@@author}}
\def\IACR@@@author[#1]#2{\authorrunning{#1}\gdef\@author{#2}}
\def\IACR@@author#1{\gdef\@author{#1}}

\if@anonymous
\gdef\@author{Anonymous Submission to \publnameshort}
\renewcommand{\author}[2][]{}
\renewcommand{\authorrunning}[1]{}
\renewcommand{\institute}[2][]{}
\fi


\def\title{\@ifnextchar[{\IACR@@@title}{\IACR@@title}}
\def\IACR@@@title[#1]#2{\gdef\@title{#2}\titlerunning{#1}}
\def\IACR@@title#1{\gdef\@title{#1}}

\newcommand{\subtitle}[1]{\gdef\@subtitle{#1}}

\newcommand{\inst}[1]{\unskip$^{#1}$}
\def\fnmsep{\unskip$^,$}


% Head/foot
\RequirePackage{fancyhdr}
\RequirePackage{graphicx}

\if@submission
\else
  \if@preprint
  \else
    % recent versions of hyperxmp load totpages, which conflicts with lastpage
    \@ifpackageloaded{totpages}{
      \def\IACR@lp{\if@loadhr\pageref*{TotPages}\else\pageref{TotPages}\fi}
    }{
      \if@IACR@lastpage
        \RequirePackage{lastpage}
      \else
        \typeout{IACR Trans: not loading lastpage, please use \@backslashchar setlastpage}
      \fi%!lastpage
    }
  \fi%!preprint
\fi%!submission

\fancypagestyle{title}{%
\fancyhf{} % clear all header and footer fields
\if@submission
\else
  \if@preprint
  \else
    \fancyhead[L]{%
      \small%
      \publname{}\\
      ISSN~\IACR@ISSN, Vol.~\IACR@vol, No.~\IACR@no, pp.~\IACR@fp--\IACR@lp. \hfill{}%
      \if@loadhr{\href{https://doi.org/\IACR@DOI}{DOI:\IACR@DOI}}\else{DOI:\IACR@DOI}\fi%
    }
    \fancyfoot[L]{%
      \small%
      Licensed under %
      \if@loadhr{\href{http://creativecommons.org/licenses/by/4.0/}{Creative Commons License CC-BY 4.0.}}%
      \else{Creative Commons License CC-BY 4.0.}%
      \fi%
      \hfill{}%
      \includegraphics[clip,height=2ex]{CC-by}\\[.1em]%
      \if@IACR@Received Received: \IACR@Received \hfill{} \fi%
      \if@IACR@Revised Revised: \IACR@Revised \hfill{} \fi%
      \if@IACR@Accepted Accepted: \IACR@Accepted \hfill{} \fi%
      \if@IACR@Published Published: \IACR@Published \fi%
    }%
    \if@loadhr
      \hypersetup{pdfcopyright={Licensed under Creative Commons License CC-BY 4.0.}}
      \hypersetup{pdflicenseurl={http://creativecommons.org/licenses/by/4.0/}}
      \hypersetup{pdfsubject={\publname{}, DOI:\IACR@DOI}}
    \fi
  \fi%!preprint
\fi%!submission

\renewcommand{\headrulewidth}{0pt}
\renewcommand{\footrulewidth}{0pt}
}%fancypagestyle

\fancyhf{}
\fancyhead[RO,LE]{\thepage}
\fancyhead[RE]{%
  \ifdefined\IACR@runningtitle\IACR@runningtitle%
  \else%
    \def\thanks##1{}%
    \def\fnmsep{}%
    \def\\{}%
    \def\footnote##1{}%
    \@title%
  \fi}
\fancyhead[LO]{%
  \ifdefined\IACR@runningauthors\IACR@runningauthors%
  \else%
    \def\thanks##1{}%
    \def\inst##1{}%
    \def\fnmsep{}%
    \def\\{}%
    \def\footnote##1{}%
    \setcounter{IACR@author@cnt}{1}%
    \def\and{\stepcounter{IACR@author@cnt}%
      \ifnum\theIACR@author@cnt=\IACR@author@last\unskip\space and \ignorespaces \else\unskip, \ignorespaces\fi}
    \@author%
  \fi}

\renewcommand{\markboth}[2]{}
\pagestyle{fancy}

\def\subtitle#1{\gdef\@subtitle{#1}}

%Abstract style, keywords
\def\@IACR@keywords{No keywords given.}

\def\keywords{\@ifnextchar[{\IACR@@@keywords}{\IACR@@keywords}}
\def\IACR@@@keywords[#1]#2{\gdef\@IACR@PDFkeywords{#1}\gdef\@IACR@keywords{#2}}
\def\IACR@@keywords#1{\gdef\@IACR@keywords{#1}}

\renewenvironment{abstract}{%
  \small\quotation\setlength{\parindent}{0pt}\noindent
  \textbf{\textsf{Abstract.}}}
  {\smallskip\par\textbf{\textsf{Keywords:}}
    \def\and{\unskip\space\textperiodcentered\space\ignorespaces}\@IACR@keywords
  \endquotation%
  \if@loadhr
    %% PDF keywords
    \def\and{, }%
    \def\thanks##1{}%
    \def\footnote##1{}%
    \def\inst##1{}%
    \def\fnmsep{}%
    \def\\{}%
    \def\zap@first@space ##1{##1}
    \def\insert@last@space##1,##2{%
      ##1%
      \ifx##2\@empty\space\else, \expandafter\insert@last@space##2\fi}
    \def\zap@comma@space##1 ,##2{%
      ##1%
      \ifx##2\@empty\else, \expandafter\zap@comma@space##2\fi}
    \def\zap@dbl@space##1 ##2{%
      ##1%
      \ifx##2\@empty\else\space\expandafter\zap@dbl@space##2\fi}
    \ifdefined\@IACR@PDFkeywords
      \hypersetup{pdfkeywords=\@IACR@PDFkeywords}
    \else
      \protected@edef\@tmp{\expandafter\@IACR@keywords}
      \protected@edef\@tmp{\expandafter\insert@last@space\@tmp,\@empty}
      \protected@edef\@tmp{\expandafter\zap@comma@space\@tmp ,\@empty}
      \protected@edef\@tmp{\expandafter\insert@last@space\@tmp,\@empty}
      \protected@edef\@tmp{\expandafter\zap@dbl@space\@tmp \@empty}
      \protected@edef\@tmp{\expandafter\zap@first@space \@tmp}
      \hypersetup{pdfkeywords=\@tmp}
    \fi
    %% PDF author
    \def\zap@one##1,##2{##1}
    \def\zap@last##1,##2{\ifx##1\@empty\else\space and \expandafter\zap@one##1\fi}
    \def\zap@last@comma##1,##2,##3{%
      ##1%
      \ifx##3\@empty%
      \expandafter\zap@last\else
      ,\expandafter\zap@last@comma\fi%
      ##2,##3}
    \ifdefined\IACR@runningauthors
      \hypersetup{pdfauthor=\IACR@runningauthors}
      \typeout{IACR@AUTHOR: \IACR@runningauthors}
    \else
      \protected@edef\@tmp{\expandafter\@author}
      \protected@edef\@tmp{\expandafter\insert@last@space\@tmp,\@empty}
      \protected@edef\@tmp{\expandafter\zap@comma@space\@tmp ,\@empty}
      \protected@edef\@tmp{\expandafter\insert@last@space\@tmp,\@empty}
      \protected@edef\@tmp{\expandafter\zap@dbl@space\@tmp \@empty}
      \ifx\@tmp\empty\else
        \protected@edef\@tmp{\expandafter\zap@first@space \@tmp}
        \typeout{IACR@AUTHOR: \@tmp}
        \hypersetup{pdfauthor=\@tmp}
      \fi
    \fi
    %% PDF title
    \ifdefined\IACR@runningtitle
      \hypersetup{pdftitle=\IACR@runningtitle}
      \typeout{IACR@TITLE: \IACR@runningtitle^^J}
    \else
      \protected@edef\@tmp{\expandafter\@title}
      \protected@edef\@tmp{\expandafter\insert@last@space\@tmp,\@empty}
      \protected@edef\@tmp{\expandafter\zap@dbl@space\@tmp \@empty}
      \ifx\@tmp\empty\else
        \protected@edef\@tmp{\expandafter\zap@first@space \@tmp}
        \hypersetup{pdftitle=\@tmp}
        \typeout{IACR@TITLE: \@tmp^^J}
      \fi
    \fi
    % PDF metadata
    \if@submission\else
      \if@preprint\else
        \if@hyperxmp@doi{
          \hypersetup{%
            pdfdoi=\IACR@DOI,%
            pdfissn=\IACR@ISSN,%
            pdfpubtype=journal,%
            pdfpublication=\publname,%
            pdfvolumenum=\IACR@vol,%
            pdfissuenum=\IACR@no,%
            pdfpagerange={\IACR@fp-\IACR@lp},%
        }}
      \fi
    \fi
  \fi
}


% autoref: capitals for Sections, and adding Algorithm
\def\equationautorefname{Equation}%
\def\footnoteautorefname{footnote}%
\def\itemautorefname{item}%
\def\figureautorefname{Figure}%
\def\tableautorefname{Table}%
\def\partautorefname{Part}%
\def\appendixautorefname{Appendix}%
\def\chapterautorefname{Chapter}%
\def\sectionautorefname{Section}%
\def\subsectionautorefname{Subsection}%
\def\subsubsectionautorefname{Subsubsection}%
\def\paragraphautorefname{paragraph}%
\def\subparagraphautorefname{subparagraph}%
\def\FancyVerbLineautorefname{line}%
\def\theoremautorefname{Theorem}%
\def\pageautorefname{page}%

\def\algorithmautorefname{Algorithm}

\def\definitionautorefname{Definition}
\def\exampleautorefname{Example}
\def\exerciseautorefname{Exercise}
\def\propertyautorefname{Property}
\def\questionautorefname{Question}
\def\solutionautorefname{Solution}
\def\propositionautorefname{Proposition}
\def\problemautorefname{Problem}
\def\lemmaautorefname{Lemma}
\def\conjectureautorefname{Conjecture}
\def\corollaryautorefname{Corollary}
\def\claimautorefname{Claim}
\def\remarkautorefname{Remark}
\def\noteautorefname{Note}
\def\caseautorefname{Case}

% AMS math
\RequirePackage{amsmath,amssymb,amsthm}
\RequirePackage{mathtools}
\theoremstyle{definition}
\newtheorem{definition}{Definition}
\newtheorem{example}{Example}
\newtheorem{exercise}{Exercise}
\newtheorem{property}{Property}
\newtheorem{question}{Question}
\newtheorem{solution}{Solution}

\theoremstyle{plain}
\newtheorem{theorem}{Theorem}
\newtheorem{proposition}{Proposition}
\newtheorem{problem}{Problem}
\newtheorem{lemma}{Lemma}
\newtheorem{conjecture}{Conjecture}
\newtheorem{corollary}{Corollary}
\newtheorem*{claim}{Claim}

\theoremstyle{remark}
\newtheorem{remark}{Remark}
\newtheorem{note}{Note}
\newtheorem{case}{Case}

\theoremstyle{plain}

%Emulate LLNCS spnewtheorem
\if@spthm
\def\spnewtheorem{\@ifstar{\IACR@spstar}{\IACR@sp}}
\def\IACR@spstar#1#2#3#4{%
  \expandafter\def\csname th@#1\endcsname{\thm@headfont{#3}#4}\thm@style{#1}
  \newtheorem*{#1}{#2}}
\def\IACR@sp#1{\@ifnextchar[{\IACR@sp@b{#1}}{\IACR@sp@a{#1}}}
\def\IACR@sp@a#1#2{%
  \@ifnextchar[{\IACR@sp@ab{#1}{#2}}{\IACR@sp@aa{#1}{#2}}}
\def\IACR@sp@ab#1#2[#3]#4#5{%
 \expandafter\def\csname th@#1\endcsname{\thm@headfont{#4}#5}\thm@style{#1}
  \newtheorem{#1}{#2}[#3]}
\def\IACR@sp@aa#1#2#3#4{%
 \expandafter\def\csname th@#1\endcsname{\thm@headfont{#3}#4}\thm@style{#1}
  \newtheorem{#1}{#2}}
\def\IACR@sp@b#1[#2]#3#4#5{%
 \expandafter\def\csname th@#1\endcsname{\thm@headfont{#4}#5}\thm@style{#1}
  \newtheorem{#1}[#2]{#3}}
\let\real@proof\proof
\def\proof{\@ifnextchar[{\proof@sptm}{\proof@@sptm}}
\def\proof@sptm[#1]{\real@proof[\proofname{} (#1)]}
\def\proof@@sptm{\real@proof}
\let\real@pushQED\pushQED
\let\real@qed\qed
\def\pushQED#1{\real@pushQED{\real@qed}}
\let\qed\qedhere
\fi
\theoremstyle{plain} %back to default

% Floats and captions
\if@floatrow
\RequirePackage{floatrow}
\floatsetup[table]{style=Plaintop}
\RequirePackage{caption}
\captionsetup{labelfont={sf,bf}}
\else
\RequirePackage{float}
\newcommand\fs@iacrabove{%
  % Swap \abovecaptionskip and \belowcaptionskip
  \addtolength\abovecaptionskip{-\belowcaptionskip}
  \addtolength\belowcaptionskip{\abovecaptionskip}
  \addtolength\abovecaptionskip{-\belowcaptionskip}
  \setlength\abovecaptionskip{-\abovecaptionskip}
  \fs@plaintop%
  \def\@fs@cfont{\sffamily\bfseries}}
\newcommand\fs@iacrbelow{%
  \fs@plain%
  \def\@fs@cfont{\sffamily\bfseries}}
\floatstyle{iacrabove}
\restylefloat{table}
\floatstyle{iacrbelow}
\restylefloat{figure}
\fi

% Extra commands
\def\email{\@ifnextchar[{\IACR@@email}{\IACR@email}}
\if@loadhr
\def\IACR@@email[#1]#2{\href{mailto:#1}{\nolinkurl{#2}}}
\def\IACR@email#1{\href{mailto:#1}{\nolinkurl{#1}}}
\else
\RequirePackage{url}
\def\IACR@@email[#1]#2{\url{#2}}
\def\IACR@email#1{\url{#1}}
\fi

% Line # for submission
\newcommand\linenomathWithnumbersforAMS{%
  \ifLineNumbers
%%  \ifx\@@par\@@@par\else
    \ifnum\interlinepenalty>-\linenopenaltypar
      \global\holdinginserts\thr@@
      \advance\interlinepenalty \linenopenalty
     \ifhmode                                   % v4.3
      \advance\predisplaypenalty \linenopenalty
     \fi
%%      \advance\postdisplaypenalty \linenopenalty
      \advance\interdisplaylinepenalty \linenopenalty
    \fi
  \fi
  \ignorespaces
  }

\if@submission
\RequirePackage[mathlines]{lineno}
\RequirePackage{xcolor}
\linenumbers
\def\linenumberfont{\normalfont\tiny\sffamily\color{gray}}

% Taken from http://phaseportrait.blogspot.fr/2007/08/lineno-and-amsmath-compatibility.html
\newcommand*\patchAmsMathEnvironmentForLineno[1]{%
  \expandafter\let\csname old#1\expandafter\endcsname\csname #1\endcsname
  \expandafter\let\csname oldend#1\expandafter\endcsname\csname end#1\endcsname
  \renewenvironment{#1}%
     {\linenomathWithnumbersforAMS\csname old#1\endcsname}%
     {\csname oldend#1\endcsname\endlinenomath}}%
\newcommand*\patchBothAmsMathEnvironmentsForLineno[1]{%
  \patchAmsMathEnvironmentForLineno{#1}%
  \patchAmsMathEnvironmentForLineno{#1*}}%
\AtBeginDocument{%
\patchAmsMathEnvironmentForLineno{equation*}%
\patchBothAmsMathEnvironmentsForLineno{align}%
\patchBothAmsMathEnvironmentsForLineno{flalign}%
\patchBothAmsMathEnvironmentsForLineno{alignat}%
\patchBothAmsMathEnvironmentsForLineno{gather}%
\patchBothAmsMathEnvironmentsForLineno{multline}%
}
\fi

% Microtype
\RequirePackage{microtype}

% Fonts
\RequirePackage[T1]{fontenc}
\RequirePackage{lmodern}

\endinput
%end of file iacrtrans.cls
