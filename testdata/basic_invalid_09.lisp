;;error:5:21-25:expected bool, found u16
(module m1)
(defcolumns (X :i16))

(defconstraint c () m1.X)
