;;error:5:1-34:symbol p1 already declared
(defcolumns (P :binary@prove) (Q :binary@prove))
;; Multiple perspectives of same name
(defperspective p1 P ((B :byte)))
(defperspective p1 Q ((C :byte)))
