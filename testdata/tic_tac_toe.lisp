;; Game of tic-tac-toe
;; game representation
;; - X is 1
;; - O is 2
;; board is an array of size 9, counting from left to right, top to bottom like so
;; 1 | 2 | 3
;; 4 | 5 | 6
;; 7 | 8 | 9


;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;; Columns
;;;;;;;;;;;;;;;;;;;;;;;;;;

(defcolumns
    (BOXES :i3 :array [9])
    (STAMP :i1)
)

;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;; Shorthands
;;;;;;;;;;;;;;;;;;;;;;;;;;

(defun (sum-boxes)
       (reduce + (for i [9] [BOXES i]))
)

(defun (sum-boxes-next)
       (reduce + (for i [9] (next [BOXES i])))
)

(defun (multiply-boxes)
       (reduce * (for i [9] [BOXES i]))
)

(defun (diff-all)
       (reduce + (for i [9] (- [BOXES i] (prev [BOXES i]))))
)

(defun (diff-all-next)
       (reduce + (for i [9] (- (next [BOXES i]) [BOXES i])))
)

(defun (sum-column-1)
 (+ (+ [BOXES 1] [BOXES 4]) [BOXES 7])
)

(defun (sum-column-2)
 (+ (+ [BOXES 2] [BOXES 5]) [BOXES 8])
)

(defun (sum-column-3)
 (+ (+ [BOXES 3] [BOXES 6]) [BOXES 9])
)

(defun (sum-row-1)
       (reduce + (for i [1:3] [BOXES i]))
)

(defun (sum-row-2)
       (reduce + (for i [4:6] [BOXES i]))
)

(defun (sum-row-3)
       (reduce + (for i [7:9] [BOXES i]))
)

(defun (sum-diagonal-left-right)
 (+ (+ [BOXES 1] [BOXES 5]) [BOXES 9])
)

(defun (sum-diagonal-right-left)
 (+ (+ [BOXES 3] [BOXES 5]) [BOXES 7])
)

(defun (multiply-column-1)
 (* (* [BOXES 1] [BOXES 4]) [BOXES 7])
)

(defun (multiply-column-2)
 (* (* [BOXES 2] [BOXES 5]) [BOXES 8])
)

(defun (multiply-column-3)
 (* (* [BOXES 3] [BOXES 6]) [BOXES 9])
)

(defun (multiply-row-1)
       (reduce * (for i [1:3] [BOXES i]))
)

(defun (multiply-row-2)
       (reduce * (for i [4:6] [BOXES i]))
)

(defun (multiply-row-3)
       (reduce * (for i [7:9] [BOXES i]))
)

(defun (multiply-diagonal-left-right)
 (* (* [BOXES 1] [BOXES 5]) [BOXES 9])
)

(defun (multiply-diagonal-right-left)
 (* (* [BOXES 3] [BOXES 5]) [BOXES 7])
)

;; function to check if the game is won or draw
;; game is won if sum is 3 and multiplication is 1 (to discard [0,1,2] combination)
;; game is won if sum is 6
;; game is a draw if sum of all boxes is 13 (if player X(1) started) or 14 (if player O(2) started)

(defun (check-column-1)
        (if (== 3 (sum-column-1))
        (if (== 1 (multiply-column-1)) 0 1)
        (if (== 6 (sum-column-1)) 0 1))
)

(defun (check-column-2)
        (if (== 3 (sum-column-2))
        (if (== 1 (multiply-column-2)) 0 1)
        (if (== 6 (sum-column-2)) 0 1))
)

(defun (check-column-3)
        (if (== 3 (sum-column-3))
        (if (== 1 (multiply-column-3)) 0 1)
        (if (== 6 (sum-column-3)) 0 1))
)
(defun (check-row-1)
        (if (== 3 (sum-row-1))
        (if (== 1 (multiply-row-1)) 0 1)
        (if (== 6 (sum-row-1)) 0 1))
)
(defun (check-row-2)
        (if (== 3 (sum-row-2))
        (if (== 1 (multiply-row-2)) 0 1)
        (if (== 6 (sum-row-2)) 0 1))
)
(defun (check-row-3)
        (if (== 3 (sum-row-3))
        (if (== 1 (multiply-row-3)) 0 1)
        (if (== 6 (sum-row-3)) 0 1))
)
(defun (check-diagonal-left-right)
        (if (== 3 (sum-diagonal-left-right))
        (if (== 1 (multiply-diagonal-left-right)) 0 1)
        (if (== 6 (sum-diagonal-left-right)) 0 1))
)
(defun (check-diagonal-right-left)
        (if (== 3 (sum-diagonal-right-left))
        (if (== 1 (multiply-diagonal-right-left)) 0 1)
        (if (== 6 (sum-diagonal-right-left)) 0 1))
)

;; check-win-or-draw returns
;; 0 if the game is won or if it's a draw
;; else 1
(defun (check-win-or-draw-version2)
 (* (* (* (* (* (* (* (* (* (* (check-column-1) (check-column-2))
                    (check-column-3))
                    (check-row-1))
                    (check-row-2))
                    (check-row-3))
                    (check-diagonal-left-right))
                    (check-diagonal-right-left))
                    (- 13 (sum-boxes)))
                    (- 14 (sum-boxes))))
)

;; this version might be les verbose or more readable than the previous one
;; TODO: and! not working with command line
(defun (check-win-or-draw)
        (if (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨ (∨
            (∧ (== 3 (sum-column-1)) (== 1 (multiply-column-1)))
            (== 6 (sum-column-1)))
            (∧ (== 3 (sum-column-2)) (== 1 (multiply-column-2))))
            (== 6 (sum-column-2)))
            (∧ (== 3 (sum-column-3)) (== 1 (multiply-column-3))))
            (== 6 (sum-column-3)))
            (∧ (== 3 (sum-row-1)) (== 1 (multiply-row-1))))
            (== 6 (sum-row-1)))
            (∧ (== 3 (sum-row-2)) (== 1 (multiply-row-2))))
            (== 6 (sum-row-2)))
            (∧ (== 3 (sum-row-3)) (== 1 (multiply-row-3))))
            (== 6 (sum-row-3)))
            (∧ (== 3 (sum-diagonal-left-right)) (== 1 (multiply-diagonal-left-right))))
            (== 6 (sum-diagonal-left-right)))
            (∧ (== 3 (sum-diagonal-right-left)) (== 1 (multiply-diagonal-right-left))))
            (== 6 (sum-diagonal-right-left)))
            (== 13 (sum-boxes)))
            (== 14 (sum-boxes))) 0 1)
)

;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;; Constraints
;;;;;;;;;;;;;;;;;;;;;;;;;;

;; range check
;; this constraint enforces that BOXES in the range [0, 1, 2]
(definrange  [BOXES 1]   3)
(definrange  [BOXES 2]   3)
(definrange  [BOXES 3]   3)
(definrange  [BOXES 4]   3)
(definrange  [BOXES 5]   3)
(definrange  [BOXES 6]   3)
(definrange  [BOXES 7]   3)
(definrange  [BOXES 8]   3)
(definrange  [BOXES 9]   3)

;; constaints for STAMP

(defconstraint   stamp-initially-vanishes (:domain {0})
                 (== STAMP 0))

(defconstraint   stamp-is-non-decreasing ()
                 (if (!= STAMP 0)
                    (== (next STAMP) 1)))


;; Game constraints

;; a player cannot play twice in a row
;; the diff-all cannot remain constant, such as (1 -> 1) or (2 ->2), it has to alternate
(defconstraint players-take-turns
                    ()
                    (begin
                        (if (== (diff-all) 1) (== (diff-all-next) 2))
                        (if (== (diff-all) 2) (== (diff-all-next) 1))
                    )
)

;; a player cannot change the value of a box if it's already played
;; a non-zero box cannot be changed
(defconstraint player-plays-in-empty-box
                 ()
                 (begin
                    (if (!= [BOXES 1] 0) (will-remain-constant! [BOXES 1]))
                    (if (!= [BOXES 2] 0) (will-remain-constant! [BOXES 2]))
                    (if (!= [BOXES 3] 0) (will-remain-constant! [BOXES 3]))
                    (if (!= [BOXES 4] 0) (will-remain-constant! [BOXES 4]))
                    (if (!= [BOXES 5] 0) (will-remain-constant! [BOXES 5]))
                    (if (!= [BOXES 6] 0) (will-remain-constant! [BOXES 6]))
                    (if (!= [BOXES 7] 0) (will-remain-constant! [BOXES 7]))
                    (if (!= [BOXES 8] 0) (will-remain-constant! [BOXES 8]))
                    (if (!= [BOXES 9] 0) (will-remain-constant! [BOXES 9]))
                 )
)

;; game stops when a player has won
;; check on STAMP to avoid testing row 0
(defconstraint game-stops-after-win-or-draw
                 ()
                 (if (!= STAMP 0)
                    (if (== (check-win-or-draw) 0)
                            (== (sum-boxes-next) 0))
                 )
)


;; last row has to be either a win, a draw or an empty row
;; should fail if game is mid-way
(defconstraint game-stops-with-win-or-draw
                 (:domain {-1})
                 (== (check-win-or-draw) 0)
)

