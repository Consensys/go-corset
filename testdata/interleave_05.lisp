(defcolumns (P :binary@loob) (X :i16@loob))
(defperspective p1 P ((Y :i16@loob)))
(definterleaved Z (X p1/Y))
(defconstraint c1 () Z)
