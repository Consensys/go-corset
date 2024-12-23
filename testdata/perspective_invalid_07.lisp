;;error:7:33-35:unknown symbol
(defpurefun ((vanishes! :@loob :force) e0) e0)
;;
(defcolumns A (P :binary@prove))
(defperspective p1 P ((B :byte)))
;; Invalid perspective name
(defconstraint c1 (:perspective p2) (vanishes! (- A B)))
