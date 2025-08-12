;;error:2:30-31:invalid padding value
(defcolumns (X :i16 :padding x))
(defconstraint c1 () (!= 0 X))
