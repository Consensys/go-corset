;;error:14:53-57:unknown symbol
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
(defconstraint c1 (:perspective p1) (vanishes! (- A p2/B)))

;; Section 2
(defperspective p2 Q ((C :binary)))
