;;error:6:26-30:qualified access not permitted here
;;error:6:32-33:expected bool, found u16
(module m1)
(defcolumns (X :i16))

(defconstraint c (:guard m1.X) X)
