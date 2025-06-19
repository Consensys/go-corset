;;error:5:26-27:unknown symbol
;;error:5:29-30:expected bool, found u16
(defcolumns (X :i16))

(defconstraint c (:guard Y) X)
