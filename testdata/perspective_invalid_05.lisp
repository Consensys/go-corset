(defcolumns (P :binary@prove) (Q :binary@prove))
;; Multiple perspectives of same name
(defperspective p1 P ((B :byte)))
(defperspective p1 Q ((C :byte)))
