(defcolumns X Y ST)
(defconstraint c1 () (* ST (+ (shift X 1) Y)))
