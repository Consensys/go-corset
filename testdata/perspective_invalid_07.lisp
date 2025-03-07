;;error:7:33-35:unknown symbol
(defpurefun ((vanishes! :𝔽@loob :force) x) x)
;;
(defcolumns A (P :binary@prove))
(defperspective p1 P ((B :byte)))
;; Invalid perspective name
(defconstraint c1 (:perspective p2) (vanishes! (- A B)))
