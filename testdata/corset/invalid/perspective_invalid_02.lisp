;;error:15:43-44:unknown symbol
;;error:16:43-44:unknown symbol
;;
;;
(defcolumns
    ;; Column (not in perspective)
    (A :i16)
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

(defperspective p1 P ((B :byte)))
(defperspective p2 Q ((C :byte)))
(defconstraint c1 (:perspective p1) (== A C))
(defconstraint c2 (:perspective p2) (== A B))
