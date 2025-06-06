;;
;;
(defcolumns
    ;; Column (not in perspective)
    (A :i16)
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

;; Section 1
(defperspective p1 P ((B :binary)))
(defconstraint c1 (:perspective p1) (== A p2/C))

;; Section 2
(defperspective p2 Q ((C :binary)))
(defconstraint c2 (:perspective p2) (== 0 (* A p1/B)))
