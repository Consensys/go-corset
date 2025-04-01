;;

(module m1)
(defcolumns (A :i16) (P :binary@prove))
(defperspective p1 P ((B :binary)))
(defconstraint c1 (:perspective p1) (== A B))

(module m2)
(defcolumns (A :i32) (P :binary@prove))
(defperspective p2 P ((B :binary)))
(defconstraint c1 (:perspective p2) (== 0 (* A B)))
