;;error:7:33-35:unknown symbol
;;
;;
(defcolumns A (P :binary@prove))
(defperspective p1 P ((B :byte)))
;; Invalid perspective name
(defconstraint c1 (:perspective p2) (== A B))
