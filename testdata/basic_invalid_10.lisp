;;error:5:32-33:expected bool, found u16
(module m1)
(defcolumns (X :i16))

(defconstraint c (:guard m1.X) X)
