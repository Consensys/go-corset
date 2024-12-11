(defcolumns (X :@loob) Y Z)
(defconstraint c1 () (if X (begin Y Z)))
