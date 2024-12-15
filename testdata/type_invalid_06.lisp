;;error:2:1-2:blah
(defcolumns (BIT :i1@loob) (X :i1@loob))

(defconstraint c1 (:guard BIT) X)
