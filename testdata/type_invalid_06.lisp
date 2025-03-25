;;error:4:32-33:expected bool, found u1
(defcolumns (BIT :i1) (X :i1))

(defconstraint c1 (:guard BIT) X)
