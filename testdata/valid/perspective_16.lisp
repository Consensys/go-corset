;;
(defcolumns
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

;; Section 1
(defperspective p1 P ((A :binary) (B :binary)))
(defconstraint c1 (:perspective p1) (== A B))

;; Section 2
(defperspective p2 Q ((C :binary) (D :byte)))
(defconstraint c2 (:perspective p2) (== 0 (* C D)))
