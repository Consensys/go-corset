;;error:4:23-32:symbol A already declared
(defcolumns A (P :binary@prove))
;; attempt to redeclare column
(defperspective p1 P ((A :byte)))
