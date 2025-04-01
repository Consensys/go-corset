(defcolumns (P :binary) (X :i16))
(defperspective p1 P ((Y :i16)))
(definterleaved Z (X p1/Y))
(defconstraint c1 () (== 0 Z))
