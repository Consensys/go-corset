;;error:4:37-38:not permitted in pure context
;;error:4:22-40:expected bool, found u32
(defcolumns (X :i16) (ST :i16))
(defconstraint c1 () (* ST (shift X X)))
