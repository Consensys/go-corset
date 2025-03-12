;;error:5:27-28:conflicting context
;;error:5:22-29:expected loobean constraint (found u16)
(defcolumns (X :i16) (Y :i16))
(definterleaved A (X Y))
(defconstraint c1 () (+ A X))
