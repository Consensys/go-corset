;;error:2:21-29:missing padding value
(defcolumns (X :i16 :padding))
(defconstraint c1 () (!= 0 X))
