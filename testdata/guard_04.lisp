(defcolumns ST (A :@loob) B C)
(defconstraint c1 (:guard ST) (if A B C))
