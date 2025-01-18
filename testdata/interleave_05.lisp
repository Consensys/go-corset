(defcolumns (P :binary@loob) (X :@loob))
(defperspective p1 P ((Y :@loob)))
(definterleaved Z (X p1/Y))
(defconstraint c1 () Z)
